# Terrarium

Mutexed file sync backed by S3. Designed for painless state management in CI or other ephemeral filesystems.

## Usage

Here is an example of persisting [Terraform](https://www.terraform.io/) state in a S3 bucket:

```
# Security credentials must be set as environment variables.
export AWS_ACCESS_KEY="..."
export AWS_SECRET_KEY="..."

# Lock and download the 'terraform-nginx' directory in the 'us-east-1'
# region bucket'my-bucket' S3 for a maximum of 10 minutes.
terrarium -expires=600 \
  -s3-bucket=my-bucket \
  -s3-region=us-east-1 \
  lock terraform-nginx

# Work with the state locally within the 'terraform-nginx/state/' directory.
terraform apply \
  -auto-approve \
  -state=terraform-nginx/state/terraform.tfstate

# Upload state and unlock the directory in S3.
terrarium -s3-bucket=my-bucket \
  -s3-region=us-east-1 \
  commit terraform-nginx
```
