import boto3
import time
import json
import logging
import os
import random
import string
import sys
import subprocess

# ==============================================================================
# --- Configuration
# ==============================================================================
# The base name for the S3 bucket. A random suffix will be added.
# IMPORTANT: You may need to change this if the base name is not compliant with
# S3 bucket naming rules or is already in use in a way that prevents creation.
S3_BUCKET_BASE_NAME = "logguardian-e2e-test-bucket"

# The AWS Region to run the test in. Can be overridden with an environment variable.
REGION = os.environ.get("AWS_REGION", "us-east-1")

# Default retention policy to be enforced by the deployed LogGuardian.
LOG_RETENTION_DAYS = 14

# Setup logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')

class E2ETestRunner:
    """
    A class to manage the end-to-end integration test for LogGuardian.
    It handles resource creation, test execution, and cleanup.
    """

    def __init__(self, region, bucket_base_name, retention_days):
        """Initializes the test runner with necessary clients and configuration."""
        self.region = region
        self.retention_days = retention_days
        
        # Generate a unique ID for this test run to prevent resource collisions
        self.run_id = ''.join(random.choices(string.ascii_lowercase + string.digits, k=8))
        
        # Define unique names for resources for this specific test run
        self.s3_bucket_name = f"{bucket_base_name}-{self.run_id}"
        self.stack_name = f"logguardian-e2e-stack-{self.run_id}"
        self.log_group_name_non_compliant = f"/aws/lambda/logguardian-test-non-compliant-{self.run_id}"
        self.log_group_name_compliant = f"/aws/lambda/logguardian-test-compliant-{self.run_id}"
        self.log_group_name_remediate = f"/aws/lambda/logguardian-test-remediate-{self.run_id}"
        
        # Boto3 clients
        self.cf_client = boto3.client("cloudformation", region_name=self.region)
        self.s3_client = boto3.client("s3", region_name=self.region)
        self.lambda_client = boto3.client("lambda", region_name=self.region)
        self.logs_client = boto3.client("logs", region_name=self.region)

        self.lambda_function_name = None

    def run(self):
        """Executes the full test lifecycle."""
        try:
            # --- Setup Phase ---
            self._check_sam_cli()
            self._setup_s3_bucket()
            self._package_and_deploy()
            self.lambda_function_name = self._get_lambda_function_name()
            logging.info(f"Successfully deployed Lambda: {self.lambda_function_name}")

            # --- Test Execution Phase ---
            self._test_non_compliant_log_group()
            self._test_compliant_log_group()
            self._test_remediation_of_existing_policy()

            logging.info("✅ All E2E integration tests passed successfully!")

        except Exception as e:
            logging.error(f"❌ An error occurred during the integration test: {e}", exc_info=True)
            sys.exit(1) # Exit with an error code
        finally:
            # --- Cleanup Phase ---
            self._cleanup()

    def _check_sam_cli(self):
        """Checks if the AWS SAM CLI is installed."""
        logging.info("Checking for AWS SAM CLI...")
        try:
            # Use subprocess.run for safer execution
            subprocess.run(["sam", "--version"], check=True, capture_output=True, text=True)
            logging.info("SAM CLI found.")
        except (subprocess.CalledProcessError, FileNotFoundError):
            logging.error("AWS SAM CLI is not installed or not in your PATH. Please install it to run this test.")
            raise EnvironmentError("SAM CLI not found.")

    def _setup_s3_bucket(self):
        """Creates the S3 bucket required for SAM deployment."""
        logging.info(f"Creating S3 bucket '{self.s3_bucket_name}' for SAM deployment...")
        try:
            if self.region == "us-east-1":
                self.s3_client.create_bucket(Bucket=self.s3_bucket_name)
            else:
                self.s3_client.create_bucket(
                    Bucket=self.s3_bucket_name,
                    CreateBucketConfiguration={'LocationConstraint': self.region}
                )
            self.s3_client.get_waiter('bucket_exists').wait(Bucket=self.s3_bucket_name)
            logging.info(f"S3 bucket '{self.s3_bucket_name}' created successfully.")
        except self.s3_client.exceptions.ClientError as e:
            logging.error(f"Failed to create S3 bucket: {e}")
            raise

    def _package_and_deploy(self):
        """Packages the SAM template and deploys it using subprocess.run for security."""
        packaged_template_path = f"packaged-template-{self.run_id}.yaml"
        
        logging.info("Packaging SAM application...")
        try:
            subprocess.run([
                "sam", "package",
                "--template-file", "template.yaml",
                "--output-template-file", packaged_template_path,
                "--s3-bucket", self.s3_bucket_name
            ], check=True, capture_output=True, text=True)
        except subprocess.CalledProcessError as e:
            logging.error(f"'sam package' command failed.\nStdout: {e.stdout}\nStderr: {e.stderr}")
            raise RuntimeError("'sam package' command failed.")

        logging.info(f"Deploying CloudFormation stack '{self.stack_name}'...")
        try:
            subprocess.run([
                "sam", "deploy",
                "--template-file", packaged_template_path,
                "--stack-name", self.stack_name,
                "--capabilities", "CAPABILITY_IAM",
                "--parameter-overrides", f"DefaultRetentionDays={self.retention_days}",
                "--region", self.region,
                "--no-fail-on-empty-changeset"
            ], check=True, capture_output=True, text=True)
        except subprocess.CalledProcessError as e:
            logging.error(f"'sam deploy' command failed.\nStdout: {e.stdout}\nStderr: {e.stderr}")
            raise RuntimeError("'sam deploy' command failed.")

        logging.info("Waiting for stack deployment to complete...")
        try:
            waiter = self.cf_client.get_waiter('stack_create_complete')
            waiter.wait(StackName=self.stack_name, WaiterConfig={'Delay': 15, 'MaxAttempts': 40})
        except self.cf_client.exceptions.ClientError as e:
            error_code = e.response.get("Error", {}).get("Code")
            error_message = e.response.get("Error", {}).get("Message")
            if error_code == "AlreadyExistsException" or "already exists" in error_message.lower():
                logging.info("Stack already exists, waiting for update to complete...")
                waiter = self.cf_client.get_waiter('stack_update_complete')
                waiter.wait(StackName=self.stack_name, WaiterConfig={'Delay': 15, 'MaxAttempts': 40})
            elif error_code == "ValidationError" and ("No updates are to be performed" in error_message or "ROLLBACK_COMPLETE" in error_message):
                logging.warning(f"Stack '{self.stack_name}' is in a terminal state without active update: {error_message}. Proceeding.")
            else:
                raise
        logging.info(f"Stack '{self.stack_name}' deployed/updated successfully.")
        
        if os.path.exists(packaged_template_path):
            os.remove(packaged_template_path)

    def _get_lambda_function_name(self):
        """Retrieves the physical name of the deployed Lambda function from stack outputs."""
        try:
            response = self.cf_client.describe_stacks(StackName=self.stack_name)
            outputs = response["Stacks"][0]["Outputs"]
            for output in outputs:
                if output["OutputKey"] == "LogGuardianFunction":
                    return output["OutputValue"]
            raise Exception("Could not find Lambda function name in stack outputs.")
        except self.cf_client.exceptions.ClientError as e:
            logging.error(f"Failed to describe stack '{self.stack_name}': {e}")
            raise

    def _invoke_lambda(self, log_group_name, test_description):
        """Invokes the LogGuardian Lambda for a specific log group."""
        logging.info(f"Invoking LogGuardian for {test_description} on '{log_group_name}'...")
        event_payload = {
            "invokingEvent": json.dumps({
                "configurationItem": {
                    "resourceType": "AWS::Logs::LogGroup",
                    "resourceId": log_group_name,
                    "configurationItemStatus": "OK"
                }
            }),
            "ruleParameters": "{}", "resultToken": "test-token"
        }
        self.lambda_client.invoke(
            FunctionName=self.lambda_function_name,
            InvocationType='Event',  # Asynchronous
            Payload=json.dumps(event_payload)
        )

    def _poll_for_retention_policy(self, log_group_name, expected_retention, timeout=60, interval=5):
        """Polls the log group until the retention policy matches the expected value or a timeout occurs."""
        logging.info(f"Polling '{log_group_name}' for retention policy of '{expected_retention}' days...")
        end_time = time.time() + timeout
        while time.time() < end_time:
            current_retention = self._get_log_group_retention(log_group_name)
            if current_retention == expected_retention:
                logging.info(f"Polling success: Found expected retention of {current_retention} days.")
                return True
            time.sleep(interval)
        
        final_retention = self._get_log_group_retention(log_group_name)
        logging.error(
            f"Polling timeout: Expected retention '{expected_retention}', but final value was '{final_retention}' after {timeout} seconds."
        )
        return False

    def _test_non_compliant_log_group(self):
        """Test Case 1: Creates a log group with no retention policy and verifies it gets remediated."""
        logging.info("\n--- Test Case 1: Non-Compliant Log Group (No Policy) ---")
        self._create_log_group(self.log_group_name_non_compliant)
        
        initial_retention = self._get_log_group_retention(self.log_group_name_non_compliant)
        assert initial_retention is None, f"FAIL: Expected no initial retention, found {initial_retention}"
        logging.info("Verified: Log group created with no retention policy.")

        self._invoke_lambda(self.log_group_name_non_compliant, "remediation")

        remediated = self._poll_for_retention_policy(self.log_group_name_non_compliant, self.retention_days)
        
        assert remediated, f"FAIL: Log group was not remediated to {self.retention_days} days."
        logging.info(f"PASS: Log group was successfully remediated.")

    def _test_compliant_log_group(self):
        """Test Case 2: Creates a log group that is already compliant and verifies it is not changed."""
        logging.info(f"\n--- Test Case 2: Compliant Log Group (Policy Matches) ---")
        self._create_log_group(self.log_group_name_compliant, retention_days=self.retention_days)

        initial_retention = self._get_log_group_retention(self.log_group_name_compliant)
        assert initial_retention == self.retention_days, f"FAIL: Expected initial retention {self.retention_days}, found {initial_retention}"
        logging.info(f"Verified: Log group created with a compliant retention policy of {initial_retention} days.")
        
        self._invoke_lambda(self.log_group_name_compliant, "compliance check")
        
        # Wait a short period to ensure no unintended action took place
        logging.info("Waiting for 15 seconds to ensure no changes are made...")
        time.sleep(15)
        
        final_retention = self._get_log_group_retention(self.log_group_name_compliant)
        assert final_retention == self.retention_days, f"FAIL: Retention policy changed from {initial_retention} to {final_retention}"
        logging.info("PASS: Compliant log group was correctly left unchanged.")
        
    def _test_remediation_of_existing_policy(self):
        """Test Case 3: Creates a log group with a non-compliant retention policy and verifies it's updated."""
        incorrect_retention = 3
        logging.info(f"\n--- Test Case 3: Non-Compliant Log Group (Incorrect Policy) ---")
        self._create_log_group(self.log_group_name_remediate, retention_days=incorrect_retention)
        
        initial_retention = self._get_log_group_retention(self.log_group_name_remediate)
        assert initial_retention == incorrect_retention, f"FAIL: Expected initial retention {incorrect_retention}, found {initial_retention}"
        logging.info(f"Verified: Log group created with an incorrect retention policy of {initial_retention} days.")

        self._invoke_lambda(self.log_group_name_remediate, "remediation")

        remediated = self._poll_for_retention_policy(self.log_group_name_remediate, self.retention_days)

        assert remediated, f"FAIL: Log group was not remediated from {incorrect_retention} to {self.retention_days} days."
        logging.info(f"PASS: Log group was successfully remediated from {incorrect_retention} to {self.retention_days} days.")

    def _create_log_group(self, name, retention_days=None):
        """Helper to create a log group for testing."""
        logging.info(f"Creating log group: {name}")
        try:
            self.logs_client.create_log_group(logGroupName=name)
            if retention_days:
                self.logs_client.put_retention_policy(logGroupName=name, retentionInDays=retention_days)
        except self.logs_client.exceptions.ResourceAlreadyExistsException:
            logging.warning(f"Log group '{name}' already exists. Deleting and recreating for a clean test state.")
            self.logs_client.delete_log_group(logGroupName=name)
            self.logs_client.create_log_group(logGroupName=name)
            if retention_days:
                self.logs_client.put_retention_policy(logGroupName=name, retentionInDays=retention_days)

    def _get_log_group_retention(self, name):
        """Helper to get the retention policy of a log group."""
        try:
            response = self.logs_client.describe_log_groups(logGroupNamePrefix=name)
            # Find the exact match
            for lg in response.get('logGroups', []):
                if lg['logGroupName'] == name:
                    return lg.get('retentionInDays')
            return None
        except self.logs_client.exceptions.ResourceNotFoundException:
            return None
            
    def _cleanup(self):
        """Deletes all AWS resources created during the test run."""
        logging.info("\n--- Starting Cleanup ---")
        
        # List of log groups to delete
        log_groups_to_delete = [
            self.log_group_name_non_compliant,
            self.log_group_name_compliant,
            self.log_group_name_remediate
        ]
        
        for lg_name in log_groups_to_delete:
            try:
                logging.info(f"Deleting log group: {lg_name}")
                self.logs_client.delete_log_group(logGroupName=lg_name)
            except self.logs_client.exceptions.ResourceNotFoundException:
                logging.info(f"Log group '{lg_name}' already deleted.")
            except Exception as e:
                logging.warning(f"Could not delete log group '{lg_name}': {e}")
        
        try:
            logging.info(f"Deleting CloudFormation stack: {self.stack_name}")
            self.cf_client.delete_stack(StackName=self.stack_name)
            waiter = self.cf_client.get_waiter('stack_delete_complete')
            waiter.wait(StackName=self.stack_name, WaiterConfig={'Delay': 30, 'MaxAttempts': 20})
            logging.info("Stack deleted successfully.")
        except self.cf_client.exceptions.ClientError as e:
            if "does not exist" in e.response['Error']['Message']:
                logging.info("Stack already deleted.")
            else:
                logging.warning(f"Could not delete stack '{self.stack_name}': {e}")

        try:
            logging.info(f"Deleting S3 bucket: {self.s3_bucket_name}")
            # Empty the bucket before deletion
            paginator = self.s3_client.get_paginator('list_object_versions')
            for page in paginator.paginate(Bucket=self.s3_bucket_name):
                for obj in page.get('Versions', []):
                    self.s3_client.delete_object(Bucket=self.s3_bucket_name, Key=obj['Key'], VersionId=obj['VersionId'])
                for marker in page.get('DeleteMarkers', []):
                    self.s3_client.delete_object(Bucket=self.s3_bucket_name, Key=marker['Key'], VersionId=marker['VersionId'])
            self.s3_client.delete_bucket(Bucket=self.s3_bucket_name)
        except Exception as e:
            logging.warning(f"Could not delete S3 bucket '{self.s3_bucket_name}'. Please delete it manually. Error: {e}")

if __name__ == "__main__":
    logging.info("=============================================")
    logging.info("=  LogGuardian E2E Integration Test Runner  =")
    logging.info("=============================================")
    
    runner = E2ETestRunner(
        region=REGION,
        bucket_base_name=S3_BUCKET_BASE_NAME,
        retention_days=LOG_RETENTION_DAYS
    )
    runner.run()