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

output "aws_secretsmanager_secret_default_id" {
  value = aws_secretsmanager_secret.default.id
  description = "ARN of the secret."
}

output "aws_secretsmanager_secret_default_arn" {
  value = aws_secretsmanager_secret.default.arn
  description = "ARN of the secret."
}

output "aws_secretsmanager_secret_default_rotation_enabled" {
  value = aws_secretsmanager_secret.default.rotation_enabled
  description = "Whether automatic rotation is enabled for this secret."
}

output "aws_secretsmanager_secret_default_replica" {
  value = aws_secretsmanager_secret.default.replica
  description = "Attributes of a replica are described below."
}

output "aws_secretsmanager_secret_default_tags_all" {
  value = aws_secretsmanager_secret.default.tags_all
  description = <<EOF
Map of tags assigned to the resource, including those inherited from the provider
 default_tags configuration block.
EOF
}
