/*
AWS Resource: aws_secretsmanager_secret - default
Generated with love by Terrawrap, an InfraCasts, LLC tool!
https://infracasts.com

The code generated below was generated using MPL v2.0 licensed code 
and documentation, and as such is is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this file, You
can obtain one at https://mozilla.org/MPL/2.0/.

Copyright © 2022 Mo Omer <mo@infracasts.com>

*/


variable "aws_secretsmanager_secret_default_description" {
  type = string
  
  description = "Description of the secret."
}

variable "aws_secretsmanager_secret_default_kms_key_id" {
  type = string
  
  description = <<EOF
ARN or Id of the AWS KMS key to be used to encrypt the secret values in the versions
 stored in this secret. If you don't specify this value, then Secrets Manager defaults
 to using the AWS account's default KMS key (the one named aws/secretsmanager). If
 the default KMS key with that name doesn't yet exist, then AWS Secrets Manager creates
 it for you automatically the first time.
EOF
}

variable "aws_secretsmanager_secret_default_name_prefix" {
  type = string
  
  description = "Creates a unique name beginning with the specified prefix. Conflicts with name."
}

variable "aws_secretsmanager_secret_default_name" {
  type = string
  
  description = <<EOF
Friendly name of the new secret. The secret name can consist of uppercase letters,
 lowercase letters, digits, and any of the following characters: /_+=.@- Conflicts
 with name_prefix.
EOF
}

variable "aws_secretsmanager_secret_default_policy" {
  type = string
  
  description = <<EOF
Valid JSON document representing a resource policy. For more information about building
 AWS IAM policy documents with Terraform, see the AWS IAM Policy Document Guide.
 Removing policy from your configuration or setting policy to null or an empty string
 (i.e., policy = "") will not delete the policy since it could have been set by aws_secretsmanager_secret_policy.
 To delete the policy, set it to "{}" (an empty JSON document
).
EOF
}

variable "aws_secretsmanager_secret_default_recovery_window_in_days" {
  type = number
  default = 30
  description = <<EOF
Number of days that AWS Secrets Manager waits before it can delete the secret. This
 value can be 0 to force deletion without recovery or range from 7 to 30 days. The
 default value is 30.
EOF
}

variable "aws_secretsmanager_secret_default_replica" {
  type = list(any)
  
  description = "Configuration block to support secret replication. See details below."
}

variable "aws_secretsmanager_secret_default_force_overwrite_replica_secret" {
  type = bool
  default = false
  description = <<EOF
Accepts boolean value to specify whether to overwrite a secret with the same name
 in the destination Region.
EOF
}

variable "aws_secretsmanager_secret_default_rotation_lambda_arn" {
  type = string
  
  description = <<EOF
ARN of the Lambda function that can rotate the secret. Use the aws_secretsmanager_secret_rotation
 resource to manage this configuration instead. As of version 2.67.0, removal of
 this configuration will no longer remove rotation due to supporting the new resource.
 Either import the new resource and remove the configuration or manually remove
 rotation.
EOF
}

variable "aws_secretsmanager_secret_default_rotation_rules" {
  type = list(any)
  
  description = <<EOF
Configuration block for the rotation configuration of this secret. Defined below.
 Use the aws_secretsmanager_secret_rotation resource to manage this configuration
 instead. As of version 2.67.0, removal of this configuration will no longer remove
 rotation due to supporting the new resource. Either import the new resource and
 remove the configuration or manually remove rotation.
EOF
}

variable "aws_secretsmanager_secret_default_tags" {
  type = map(any)
  
  description = <<EOF
Key-value map of user-defined tags that are attached to the secret. If configured
 with a provider default_tags configuration block present, tags with matching keys
 will overwrite those defined at the provider-level.
EOF
}

