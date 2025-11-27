import boto3
import time
import json
import logging
import os
import random
import string
import sys

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
        if os.system("sam --version > nul 2>&1") != 0 and os.system("sam --version > /dev/null 2>&1") != 0:
            logging.error("AWS SAM CLI is not installed or not in your PATH. Please install it to run this test.")
            raise EnvironmentError("SAM CLI not found.")
        logging.info("SAM CLI found.")

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
        """Packages the SAM template using 'sam package' and deploys it using 'sam deploy'."""
        packaged_template_path = f"packaged-template-{self.run_id}.yaml"
        
        logging.info("Packaging SAM application...")
        package_command = (
            f"sam package --template-file template.yaml "
            f"--output-template-file {packaged_template_path} "
            f"--s3-bucket {self.s3_bucket_name}"
        )
        if os.system(package_command) != 0:
            raise RuntimeError("'sam package' command failed.")

        logging.info(f"Deploying CloudFormation stack '{self.stack_name}'...")
        deploy_command = (
            f"sam deploy --template-file {packaged_template_path} "
            f"--stack-name {self.stack_name} "
            f"--capabilities CAPABILITY_IAM "
            f"--parameter-overrides DefaultRetentionDays={self.retention_days} "
            f"--region {self.region} "
            f"--no-fail-on-empty-changeset"
        )
        if os.system(deploy_command) != 0:
            raise RuntimeError("'sam deploy' command failed.")

        logging.info("Waiting for stack deployment to complete...")
        waiter = self.cf_client.get_waiter('stack_create_complete')
        waiter.wait(StackName=self.stack_name, WaiterConfig={'Delay': 30, 'MaxAttempts': 20})
        logging.info(f"Stack '{self.stack_name}' deployed successfully.")
        
        # Clean up the local packaged template file
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

    def _invoke_lambda_and_wait(self, log_group_name, test_description):
        """Invokes the LogGuardian Lambda and waits for potential remediation."""
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
        # Give Lambda time to process
        logging.info("Waiting for potential remediation (up to 30 seconds)...")
        time.sleep(30)

    def _test_non_compliant_log_group(self):
        """Test Case 1: Creates a log group with no retention policy and verifies it gets remediated."""
        logging.info("\n--- Test Case 1: Non-Compliant Log Group (No Policy) ---")
        self._create_log_group(self.log_group_name_non_compliant)
        
        initial_retention = self._get_log_group_retention(self.log_group_name_non_compliant)
        assert initial_retention is None, f"FAIL: Expected no initial retention, found {initial_retention}"
        logging.info("Verified: Log group created with no retention policy.")

        self._invoke_lambda_and_wait(self.log_group_name_non_compliant, "remediation")

        final_retention = self._get_log_group_retention(self.log_group_name_non_compliant)
        assert final_retention == self.retention_days, f"FAIL: Expected retention {self.retention_days}, found {final_retention}"
        logging.info(f"PASS: Log group was successfully remediated to {final_retention} days.")

    def _test_compliant_log_group(self):
        """Test Case 2: Creates a log group that is already compliant and verifies it is not changed."""
        logging.info(f"\n--- Test Case 2: Compliant Log Group (Policy Matches) ---")
        self._create_log_group(self.log_group_name_compliant, retention_days=self.retention_days)

        initial_retention = self._get_log_group_retention(self.log_group_name_compliant)
        assert initial_retention == self.retention_days, f"FAIL: Expected initial retention {self.retention_days}, found {initial_retention}"
        logging.info(f"Verified: Log group created with a compliant retention policy of {initial_retention} days.")
        
        self._invoke_lambda_and_wait(self.log_group_name_compliant, "compliance check")
        
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

        self._invoke_lambda_and_wait(self.log_group_name_remediate, "remediation")

        final_retention = self._get_log_group_retention(self.log_group_name_remediate)
        assert final_retention == self.retention_days, f"FAIL: Expected retention {self.retention_days}, found {final_retention}"
        logging.info(f"PASS: Log group was successfully remediated from {incorrect_retention} to {final_retention} days.")

    def _create_log_group(self, name, retention_days=None):
        """Helper to create a log group for testing."""
        logging.info(f"Creating log group: {name}")
        try:
            self.logs_client.create_log_group(logGroupName=name)
            if retention_days:
                self.logs_client.put_retention_policy(logGroupName=name, retentionInDays=retention_days)
        except self.logs_client.exceptions.ResourceAlreadyExistsException:
            logging.warning(f"Log group '{name}' already exists. Attempting to reuse.")

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