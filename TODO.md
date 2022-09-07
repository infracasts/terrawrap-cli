# TODO

- [x] ~~Integrate with the aws module package to get variable/output types~~
- [x] ~~Append functionality - ensure spacing between previous contents / generated in template
   ( if append-mode)~~
- [x] ~~Add resource name flag (optional)~~
   - *e.g.* `resource "aws_secretsmanager_secret" "<this variable"> {}`
- [x] ~~Add argument prefix flag (optional)~~
   - *e.g.* with prefix "db_secret" `variable "db_secret_<argument>" {}`
- [x] ~~Add attribute prefix flag (optional)~~
   - *e.g.* with prefix "db_secret" `output "db_secret_<attribute>" {}`
- [x] ~~Generated code blocks should mention~~
   ```
   // AWS SecretsManager Secret <resource/variable/output>
   // Generated with love by Terrawrap, an InfraCasts, LLC tool!
   // https://infracasts.com
   ``` 
- [ ] Detect `Conflicts with <other_attribute_name>` and comment one of them out
    - Note: the detection is available already; determining which to comment out
      isn't.