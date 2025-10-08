# Email OTP Node

The Email-OTP node is a powerful node that can be used in various flows.

- Onboarding: Signup with Email -> OTP verification
- Passwordless Login: Enter Email -> OTP Verification
- User Management: Forgot Password -> OTP Verification

Email OTPs come with several security challenges:
- As OTPs are short they need protection against enumartion attacks. E.g. a 6 digit otp can be brute forced if there is no mitigration
- The email otp node should not enable to spam user with otps

The email OTP Node provides the followin functionality:

- If there is a user with this email we increase the user failed attempt counters in the email attribute. Max failed attemps limits the number of possible trials.
- If there is a user with this email and the email is locked no otp is sent
- The node fails silently. This is important to avoid attacks on user enumaration
- If there is no user in the context we can still send an OTP. In that case the max number of attempts will still be increased and fail silently once the limit is reached
- If the email does not have the verified flag, the otp node will set it to verified.

Custom configuration options:
- disableLoadUser: In that case the node does not check if a user already has this address
- maxAttempts: Limits the maximum number of failed attempts
- emailTemplate: Name of the email tempalte


Email:
If the email address is explictily set via context["email"] then this is used. If a user is available in the context and no email is set then the default email address of the user is used. If no email address can be found then the node fails. If you want to collect and test an email please use the ask-email-otp node which ask for an email and directly verifies it.



