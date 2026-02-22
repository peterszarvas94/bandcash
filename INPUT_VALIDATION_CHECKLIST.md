# Input Validation Checklist

- [x] Parse request signals/params once at handler start.
- [x] Normalize input before validation (`NormalizeText`, `NormalizeEmail`).
- [x] Validate payload structs with tags and `ValidateWithLocale`.
- [x] Return `400` for malformed payloads and invalid route/query IDs.
- [x] Return `422` for semantic validation failures.
- [x] Use localized validation messages (including `validation.email`).
- [x] Strictly validate ID-shaped inputs with `utils.IsValidID`.
- [x] Use only normalized/validated values for DB and email operations.
