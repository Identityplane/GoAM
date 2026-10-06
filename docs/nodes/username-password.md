# Username and password nodes

These nodes collect or store a local username and password. They do **not** apply to federated login (OIDC, GitHub, Telegram, passkeys): those methods never put a password in session context, and GoAM does not store one.

## Username

`askUsername` and `askUsernamePassword` reject a username that is missing or only whitespace. The value is trimmed before it is written to context.

`checkUsernameAvailable` treats an empty username the same way (not available) and sets `errorMessage` to `Username is required`.

`createUser` validates a username only when the `username` key is present in context. OIDC and mock flows that omit the key are unchanged.

## Password

Empty passwords are always rejected at the prompt (`Password is required`). Length limits are optional on the prompt and defaulted when hashing a new password.

### `custom_config`

| Option | Nodes | Default if unset |
| --- | --- | --- |
| `minLength` | `askPassword`, `createUser`, `updatePassword` | Prompt: no minimum (login still accepts short legacy passwords). Store (`createUser` / `updatePassword`): `8`. |
| `maxLength` | `askPassword`, `createUser`, `updatePassword` | Prompt: no maximum. Store: `64`. |

Set them as strings in flow YAML:

```yaml
askPassword:
  use: askPassword
  custom_config:
    minLength: "8"
    maxLength: "64"
  next:
    submitted: createUser
```

Use this on **registration** and **password change** nodes. Do **not** set `minLength` / `maxLength` on a login `askPassword`: a user whose existing hash is shorter or longer than the new policy must still be able to sign in.

`askUsernamePassword` only rejects empty fields. It does not apply min/max (typical login form). Put length limits on a following `updatePassword` or a dedicated register `askPassword` instead.

`createUser` hashes a password only when the `password` key is present and non-empty. If the key is absent (OIDC/social/mock), no password attribute is created and length policy is skipped.

## Client behaviour

Invalid input stays on the same node. The JSON API returns HTTP 200 with `prompts` and `errorMessage`. See [JSON Authentication API](../auth_api.md#4-validation-error-same-node-http-200).
