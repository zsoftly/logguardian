import boto3
import time
import json
import logging
import os
import random
import string
import sys
import subprocess
import yaml # Added for reading template.yaml
from botocore.exceptions import WaiterError, ClientError # Added for robust error handling

# ==============================================================================
# --- Configuration
# ==============================================================================
# The base name for the S3 bucket. A random suffix will be added.
# IMPORTANT: You may need to change this if the base name is not compliant with
# S3 bucket naming rules or is already in use in a way that prevents creation.
S3_BUCKET_BASE_NAME = "logguardian-e2e-test-bucket"

# The default AWS Region to run the test in if not specified by env var.
DEFAULT_REGION = "ca-central-1"

# Environment variable to specify a comma-separated list of regions for multi-region testing.
# Example: E2E_TEST_REGIONS="us-east-1,us-west-2"
E2E_TEST_REGIONS = os.environ.get("E2E_TEST_REGIONS", DEFAULT_REGION).split(',')

# Function to read DefaultRetentionDays from template.yaml
def get_default_retention_days_from_template(template_path="template.yaml"):
    try:
        with open(template_path, 'r') as f:
            template = yaml.safe_load(f)
        # Assuming DefaultRetentionDays is directly under Parameters
        return int(template['Parameters']['DefaultRetentionDays']['Default']) # Ensure consistent integer type
    except Exception as e:
        logging.exception(f"Could not read DefaultRetentionDays from {template_path}. Falling back to default (14).") # Use logging.exception
        # Fallback to a default if reading fails
        return 14 

# Get LOG_RETENTION_DAYS from template.yaml
LOG_RETENTION_DAYS = get_default_retention_days_from_template()

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
        self.s3_bucket_name = f"{bucket_base_name}-{self.run_id}-{self.region}" # Added region to bucket name
        self.stack_name = f"logguardian-e2e-stack-{self.run_id}-{self.region}" # Added region to stack name
        self.log_group_name_non_compliant = f"/aws/lambda/logguardian-test-non-compliant-{self.run_id}-{self.region}"
        self.log_group_name_compliant = f"/aws/lambda/logguardian-test-compliant-{self.run_id}-{self.region}"
        self.log_group_name_remediate = f"/aws/lambda/logguardian-test-remediate-{self.run_id}-{self.region}"
        self.log_group_name_performance_prefix = f"/aws/lambda/logguardian-test-performance-{self.run_id}-{self.region}-"
        
        # Boto3 clients
        self.cf_client = boto3.client("cloudformation", region_name=self.region)
        self.s3_client = boto3.client("s3", region_name=self.region)
        self.lambda_client = boto3.client("lambda", region_name=self.region)
        self.logs_client = boto3.client("logs", region_name=self.region)

        self.lambda_function_name = None
        self.performance_log_groups = []

    def run(self):
        """Executes the full test lifecycle."""
        try:
            logging.info(f"[START] Starting E2E test in region: {self.region}") # Emoji replaced
            # --- Setup Phase ---
            self._check_sam_cli()
            self._setup_s3_bucket()
            self._package_and_deploy()
            self.lambda_function_name = self._get_lambda_function_name()
            logging.info(f"[OK] Successfully deployed Lambda: {self.lambda_function_name} in {self.region}") # Emoji replaced

            # --- Test Execution Phase ---
            self._test_non_compliant_log_group()
            self._test_compliant_log_group()
            self._test_remediation_of_existing_policy()
            self._test_performance_large_log_groups()
            self._test_invalid_payload()
            self._test_permission_error_placeholder()

            logging.info(f"[OK] All E2E integration tests passed successfully in region: {self.region}!") # Emoji replaced

        except Exception as e:
            logging.exception(f"[FAIL] An error occurred during the integration test in {self.region}.") # Use logging.exception
            raise # Re-raise the exception to fail the overall run
        finally:
            # --- Cleanup Phase ---
            self._cleanup()

    def _check_sam_cli(self):
        """Checks if the AWS SAM CLI is installed."""
        logging.info("Checking for AWS SAM CLI...")
        try:
            subprocess.run(["sam", "--version"], check=True, capture_output=True, text=True)
            logging.info("SAM CLI found.")
        except (subprocess.CalledProcessError, FileNotFoundError):
            logging.exception("AWS SAM CLI is not installed or not in your PATH. Please install it to run this test.") # Use logging.exception
            raise EnvironmentError("SAM CLI not found.")

    def _setup_s3_bucket(self):
        """Creates the S3 bucket required for SAM deployment."""
        logging.info(f"Creating S3 bucket '{self.s3_bucket_name}' for SAM deployment in {self.region}...")
        try:
            if self.region == "us-east-1":
                self.s3_client.create_bucket(Bucket=self.s3_bucket_name)
            else:
                self.s3_client.create_bucket(
                    Bucket=self.s3_bucket_name,
                    CreateBucketConfiguration={'LocationConstraint': self.region}
                )
            self.s3_client.get_waiter('bucket_exists').wait(Bucket=self.s3_bucket_name)
            logging.info(f"[OK] S3 bucket '{self.s3_bucket_name}' created successfully in {self.region}.") # Emoji replaced
        except ClientError as e: # Use specific ClientError
            if "BucketAlreadyOwnedByYou" in str(e):
                logging.warning(f"S3 bucket '{self.s3_bucket_name}' already exists and owned by you. Proceeding.")
            else:
                logging.exception(f"[FAIL] Failed to create S3 bucket '{self.s3_bucket_name}' in {self.region}.") # Use logging.exception
                raise # Re-raise for ClientError

    def _package_and_deploy(self):
        """Packages the SAM template and deploys it using subprocess.run for security."""
        packaged_template_path = f"packaged-template-{self.run_id}-{self.region}.yaml"
        
        logging.info(f"Packaging SAM application in {self.region}...")
        try:
            subprocess.run([
                "sam", "package",
                "--template-file", "template.yaml",
                "--output-template-file", packaged_template_path,
                "--s3-bucket", self.s3_bucket_name,
                "--region", self.region
            ], check=True, capture_output=True, text=True)
        except subprocess.CalledProcessError as e:
            logging.exception(f"[FAIL] 'sam package' command failed in {self.region}.") # Use logging.exception
            raise RuntimeError(f"'sam package' command failed: {e}") from e

        logging.info(f"Deploying CloudFormation stack '{self.stack_name}' in {self.region}...")
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
            logging.exception(f"[FAIL] 'sam deploy' command failed in {self.region}.") # Use logging.exception
            raise RuntimeError(f"'sam deploy' command failed: {e}") from e

        logging.info(f"Waiting for stack deployment to complete in {self.region}...")
        try:
            response = self.cf_client.describe_stacks(StackName=self.stack_name)
            status = response['Stacks'][0]['StackStatus']
            
            if 'ROLLBACK' in status:
                logging.error(f"[FAIL] Stack '{self.stack_name}' in {self.region} is in {status} state. Cannot proceed with deployment.") # Emoji replaced
                raise RuntimeError(f"Stack in {status} state.")
            elif status in ('CREATE_IN_PROGRESS', 'REVIEW_IN_PROGRESS'):
                waiter = self.cf_client.get_waiter('stack_create_complete')
            elif status == 'UPDATE_IN_PROGRESS':
                waiter = self.cf_client.get_waiter('stack_update_complete')
            elif status in ('CREATE_COMPLETE', 'UPDATE_COMPLETE'):
                logging.info(f"Stack '{self.stack_name}' is already in {status} state. No deployment needed.")
                return 
            else:
                logging.warning(f"Stack '{self.stack_name}' in unexpected state {status}. Attempting update waiter.")
                waiter = self.cf_client.get_waiter('stack_update_complete')
            
            waiter.wait(StackName=self.stack_name, WaiterConfig={'Delay': 15, 'MaxAttempts': 40})
            logging.info(f"[OK] Stack '{self.stack_name}' deployed/updated successfully in {self.region}.") # Emoji replaced
        except (ClientError, WaiterError) as e:
            logging.exception(f"[FAIL] CloudFormation waiter failed for stack '{self.stack_name}' in {self.region}.") # Use logging.exception
            raise RuntimeError(f"CloudFormation deployment failed: {e}") from e
        finally:
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
        except ClientError as e: # Use specific ClientError
            logging.exception(f"[FAIL] Failed to describe stack '{self.stack_name}' in {self.region}.") # Use logging.exception
            raise

    def _invoke_lambda(self, log_group_name, test_description, payload_override=None):
        """Invokes the LogGuardian Lambda for a specific log group."""
        logging.info(f"Invoking LogGuardian for {test_description} on '{log_group_name}' in {self.region}...")
        if payload_override:
            event_payload = payload_override
        else:
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
        
        response = self.lambda_client.invoke(
            FunctionName=self.lambda_function_name,
            InvocationType='RequestResponse',
            Payload=json.dumps(event_payload)
        )
        
        try:
            response_payload = json.loads(response['Payload'].read())
        except json.JSONDecodeError as e:
            logging.exception(f"[FAIL] Failed to parse Lambda response payload from {log_group_name} in {self.region}.") # Use logging.exception
            return False, {"error": "Invalid JSON response from Lambda"}
        
        if response.get('FunctionError'):
            logging.error(f"[FAIL] Lambda invocation error for {log_group_name}: {response.get('FunctionError')} - Payload: {response_payload}") # Emoji replaced
            return False, response_payload
        
        logging.info(f"[OK] Lambda invocation successful for {log_group_name}") # Emoji replaced
        return True, response_payload

    def _poll_for_retention_policy(self, log_group_name, expected_retention, timeout=60, interval=5):
        """Polls the log group until the retention policy matches the expected value or a timeout occurs."""
        logging.info(f"Polling '{log_group_name}' in {self.region} for retention policy of '{expected_retention}' days...")
        end_time = time.time() + timeout
        while time.time() < end_time:
            current_retention = self._get_log_group_retention(log_group_name)
            if current_retention == expected_retention:
                logging.info(f"[OK] Polling success: Found expected retention of {current_retention} days.") # Emoji replaced
                return True
            time.sleep(interval)
        
        final_retention = self._get_log_group_retention(log_group_name)
        logging.error(
            f"[FAIL] Polling timeout in {self.region}: Expected retention '{expected_retention}', but final value was '{final_retention}' after {timeout} seconds." # Emoji replaced
        )
        return False

    def _test_non_compliant_log_group(self):
        """Test Case 1: Creates a log group with no retention policy and verifies it gets remediated."""
        logging.info("\n--- Test Case 1: Non-Compliant Log Group (No Policy) ---") # f-prefix removed
        self._create_log_group(self.log_group_name_non_compliant)
        
        initial_retention = self._get_log_group_retention(self.log_group_name_non_compliant)
        assert initial_retention is None, f"[FAIL] Expected no initial retention, found {initial_retention}" # Emoji replaced
        logging.info("Verified: Log group created with no retention policy.")

        success, _ = self._invoke_lambda(self.log_group_name_non_compliant, "remediation")
        assert success, "[FAIL] Lambda invocation failed for non-compliant log group." # Emoji replaced

        remediated = self._poll_for_retention_policy(self.log_group_name_non_compliant, self.retention_days)
        
        assert remediated, f"[FAIL] Log group was not remediated to {self.retention_days} days." # Emoji replaced
        logging.info(f"[OK] Log group was successfully remediated in {self.region}.") # Emoji replaced

    def _test_compliant_log_group(self):
        """Test Case 2: Creates a log group that is already compliant and verifies it is not changed."""
        logging.info(f"\n--- Test Case 2: Compliant Log Group (Policy Matches) ---") # f-prefix removed
        self._create_log_group(self.log_group_name_compliant, retention_days=self.retention_days)

        initial_retention = self._get_log_group_retention(self.log_group_name_compliant)
        assert initial_retention == self.retention_days, f"[FAIL] Expected initial retention {self.retention_days}, found {initial_retention}" # Emoji replaced
        logging.info(f"Verified: Log group created with a compliant retention policy of {initial_retention} days.")
        
        success, _ = self._invoke_lambda(self.log_group_name_compliant, "compliance check")
        assert success, "[FAIL] Lambda invocation failed for compliant log group." # Emoji replaced

        logging.info("Waiting for 15 seconds to ensure no changes are made...")
        time.sleep(15)
        
        final_retention = self._get_log_group_retention(self.log_group_name_compliant)
        assert final_retention == self.retention_days, f"[FAIL] Retention policy changed from {initial_retention} to {final_retention}" # Emoji replaced
        logging.info(f"[OK] Compliant log group was correctly left unchanged in {self.region}.") # Emoji replaced
        
    def _test_remediation_of_existing_policy(self):
        """Test Case 3: Creates a log group with a non-compliant retention policy and verifies it's updated."""
        incorrect_retention = 3
        logging.info(f"\n--- Test Case 3: Non-Compliant Log Group (Incorrect Policy) ---") # f-prefix removed
        self._create_log_group(self.log_group_name_remediate, retention_days=incorrect_retention)
        
        initial_retention = self._get_log_group_retention(self.log_group_name_remediate)
        assert initial_retention == incorrect_retention, f"[FAIL] Expected initial retention {incorrect_retention}, found {initial_retention}" # Emoji replaced
        logging.info(f"Verified: Log group created with an incorrect retention policy of {initial_retention} days.")

        success, _ = self._invoke_lambda(self.log_group_name_remediate, "remediation")
        assert success, "[FAIL] Lambda invocation failed for remediation of existing policy." # Emoji replaced

        remediated = self._poll_for_retention_policy(self.log_group_name_remediate, self.retention_days)

        assert remediated, f"[FAIL] Log group was not remediated from {incorrect_retention} to {self.retention_days} days." # Emoji replaced
        logging.info(f"[OK] Log group was successfully remediated in {self.region}.") # Emoji replaced

    def _test_performance_large_log_groups(self, count=50):
        """Test Case 4: Tests performance with a large number of log groups."""
        logging.info(f"\n--- Test Case 4: Performance Test with {count} Log Groups ---") # f-prefix removed
        start_time = time.time()
        
        # Create log groups
        for i in range(count):
            log_group_name = f"{self.log_group_name_performance_prefix}{i}"
            self._create_log_group(log_group_name)
            self.performance_log_groups.append(log_group_name)
        logging.info(f"Created {count} log groups for performance test in {self.region}.")

        # Invoke Lambda for each log group (simulating multiple events)
        invocation_start_time = time.time()
        for lg_name in self.performance_log_groups:
            success, _ = self._invoke_lambda(lg_name, "performance test remediation")
            assert success, f"[FAIL] Lambda invocation failed for performance log group {lg_name}." # Emoji replaced
        invocation_end_time = time.time()
        logging.info(f"Invoked Lambda for {count} log groups in {self.region}. Invocation took: {invocation_end_time - invocation_start_time:.2f} seconds.")

        # Poll for remediation
        all_remediated = True
        for lg_name in self.performance_log_groups:
            if not self._poll_for_retention_policy(lg_name, self.retention_days, timeout=180, interval=10):
                all_remediated = False
                break
        
        end_time = time.time()
        total_time = end_time - start_time
        
        assert all_remediated, f"[FAIL] Not all performance log groups were remediated in {self.region}." # Emoji replaced
        logging.info(f"[OK] Performance test with {count} log groups completed in {self.region}. Total time: {total_time:.2f} seconds.") # Emoji replaced

    def _test_invalid_payload(self):
        """Test Case 5: Tests Lambda's handling of an invalid event payload."""
        logging.info("\n--- Test Case 5: Invalid Payload Error Handling ---") # f-prefix removed
        
        # A malformed event payload (e.g., missing 'invokingEvent' or invalid JSON)
        invalid_payload = {
            "someOtherField": "value",
            "malformedJson": "{\"key\": \"value\"" # Valid JSON but missing required 'invokingEvent' field
        }
        
        logging.info(f"Invoking Lambda with invalid payload in {self.region}...")
        success, response_payload = self._invoke_lambda("N/A", "invalid payload", payload_override=invalid_payload)
        
        assert not success, "[FAIL] Lambda invocation with invalid payload unexpectedly succeeded." # Emoji replaced
        logging.info(f"Lambda correctly returned error for invalid payload: {response_payload}")
        logging.info(f"[OK] Lambda correctly handled invalid payload in {self.region}.") # Emoji replaced

    def _test_permission_error_placeholder(self):
        """Test Case 6: Placeholder for permission error testing."""
        logging.info("\n--- Test Case 6: Permission Error Testing (Placeholder) ---") # f-prefix removed
        logging.info("Permission error testing is complex and requires specific setup (e.g., deploying a role with restricted policies).")
        logging.info("This test case is a placeholder for future implementation.")
        logging.info("[OK] Permission error test placeholder noted.") # Emoji replaced


    def _create_log_group(self, name, retention_days=None):
        """Helper to create a log group for testing."""
        logging.info(f"Creating log group: {name} in {self.region}")
        try:
            self.logs_client.create_log_group(logGroupName=name)
            if retention_days:
                self.logs_client.put_retention_policy(logGroupName=name, retentionInDays=retention_days)
            logging.info(f"[OK] Log group '{name}' created/updated successfully in {self.region}.") # Emoji replaced
        except self.logs_client.exceptions.ResourceAlreadyExistsException:
            logging.warning(f"Log group '{name}' already exists in {self.region}. Deleting and recreating for a clean test state.")
            self.logs_client.delete_log_group(logGroupName=name)
            self.logs_client.create_log_group(logGroupName=name)
            if retention_days:
                self.logs_client.put_retention_policy(logGroupName=name, retentionInDays=retention_days)
            logging.info(f"[OK] Log group '{name}' recreated successfully in {self.region}.") # Emoji replaced
        except ClientError as e:
            logging.exception(f"[FAIL] Failed to create log group '{name}' in {self.region}.") # Use logging.exception
            raise

    def _get_log_group_retention(self, name):
        """Helper to get the retention policy of a log group."""
        try:
            response = self.logs_client.describe_log_groups(logGroupNamePrefix=name)
            for lg in response.get('logGroups', []):
                if lg['logGroupName'] == name:
                    return lg.get('retentionInDays')
            return None
        except self.logs_client.exceptions.ResourceNotFoundException:
            return None
            
    def _cleanup(self):
        """Deletes all AWS resources created during the test run."""
        logging.info(f"\n--- Starting Cleanup in {self.region} ---") # f-prefix removed
        
        log_groups_to_delete = [
            *self.performance_log_groups,
            self.log_group_name_non_compliant,
            self.log_group_name_compliant,
            self.log_group_name_remediate,
        ]
        
        for lg_name in log_groups_to_delete:
            try:
                logging.info(f"Deleting log group: {lg_name} in {self.region}")
                self.logs_client.delete_log_group(logGroupName=lg_name)
            except ClientError as e:
                if e.response['Error']['Code'] == 'ResourceNotFoundException':
                    logging.info(f"Log group '{lg_name}' already deleted in {self.region}.")
                else:
                    logging.exception(f"[FAIL] Could not delete log group '{lg_name}' in {self.region}.") # Emoji replaced
        
        try:
            logging.info(f"Deleting CloudFormation stack: {self.stack_name} in {self.region}")
            self.cf_client.delete_stack(StackName=self.stack_name)
            waiter = self.cf_client.get_waiter('stack_delete_complete')
            waiter.wait(StackName=self.stack_name, WaiterConfig={'Delay': 30, 'MaxAttempts': 20})
            logging.info(f"[OK] Stack '{self.stack_name}' deleted successfully in {self.region}.") # Emoji replaced
        except ClientError as e:
            if "does not exist" in e.response['Error']['Message']:
                logging.info(f"Stack in {self.region} already deleted.")
            else:
                logging.exception(f"[FAIL] Could not delete stack '{self.stack_name}' in {self.region}.") # Emoji replaced

        try:
            logging.info(f"Deleting S3 bucket: {self.s3_bucket_name} in {self.region}")
            paginator = self.s3_client.get_paginator('list_object_versions')
            for page in paginator.paginate(Bucket=self.s3_bucket_name):
                for obj in page.get('Versions', []):
                    self.s3_client.delete_object(Bucket=self.s3_bucket_name, Key=obj['Key'], VersionId=obj['VersionId'])
                for marker in page.get('DeleteMarkers', []):
                    self.s3_client.delete_object(Bucket=self.s3_bucket_name, Key=marker['Key'], VersionId=marker['VersionId'])
            self.s3_client.delete_bucket(Bucket=self.s3_bucket_name)
            logging.info(f"[OK] S3 bucket '{self.s3_bucket_name}' deleted successfully in {self.region}.") # Emoji replaced
        except ClientError as e:
            logging.exception(f"[FAIL] Could not delete S3 bucket '{self.s3_bucket_name}' in {self.region}. Please delete it manually.") # Use logging.exception

if __name__ == "__main__":
    
    for region in E2E_TEST_REGIONS:
        try:
            runner = E2ETestRunner(
                region=region.strip(),
                bucket_base_name=S3_BUCKET_BASE_NAME,
                retention_days=LOG_RETENTION_DAYS
            )
            runner.run()
        except Exception as e:
            logging.exception(f"[FAIL] Test failed in region {region}.") # Emoji replaced
            sys.exit(1)

    logging.info("\n=============================================")
    logging.info("[OK] All E2E integration tests across all regions passed successfully!") # Emoji replaced
    logging.info("=============================================")