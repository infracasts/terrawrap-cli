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

resource "aws_secretsmanager_secret" "default" {
    description = var.aws_secretsmanager_secret_default_description
    kms_key_id = var.aws_secretsmanager_secret_default_kms_key_id
    name_prefix = var.aws_secretsmanager_secret_default_name_prefix
    name = var.aws_secretsmanager_secret_default_name
    policy = var.aws_secretsmanager_secret_default_policy
    recovery_window_in_days = var.aws_secretsmanager_secret_default_recovery_window_in_days
    replica = var.aws_secretsmanager_secret_default_replica
    force_overwrite_replica_secret = var.aws_secretsmanager_secret_default_force_overwrite_replica_secret
    // rotation_lambda_arn = var.aws_secretsmanager_secret_default_rotation_lambda_arn // DEPRECATED
    // rotation_rules = var.aws_secretsmanager_secret_default_rotation_rules // DEPRECATED
    tags = var.aws_secretsmanager_secret_default_tags
}
