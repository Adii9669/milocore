# Onechat OTP delivery + production DB repair report

**Date:** 2026-11-24  
**Status:** Research only — no backend code changed in this step.  
**Scope:** Why registration OTP emails never arrive, the best delivery method to use now, whether SMS OTP is free, and the two production DB errors you also selected.

---

## 1. Executive summary

### Why no OTP arrives

The local backend checkout is `/home/adii/Projects/miloverse/backend` (branch `dev`, commit `d01b2e2`). Its code matches the Render behaviour in the logs:

1. `AuthService.Register()` creates the user as `Verified: false` but **all OTP generation, hashing, storage, and mail-send calls are commented out**. The register endpoint therefore correctly returns `201`, but no code ever exists to verify and no email is sent.
2. `NewEmailService()` has **every provider field commented out** (`client`, `apiKey`, `from`, `sendGridKey`). If `SendOTP()` were called as-is, it would dereference a nil Resend client.
3. `config.AppConfig.Email` is fully commented out, so provider keys are never parsed.
4. `/auth/resend-otp` still works logically, but it also ends at `emailService.SendOTP()`, which would fail because the email client is uninitialised.

### Your assumption about Resend/SendGrid is only half right

- **Resend default `onboarding@resend.dev` can only send to the email address that owns the Resend account.** That is a deliberate test restriction. After you **verify your own domain**, Resend can send to any recipient. This is not a dead end.
- **SendGrid free accounts can use Single Sender Verification** and send to other recipients, but SendGrid itself says Single Sender is for testing only; domain authentication is required for real production use/upgrades.
- **No legitimate provider lets you send as an arbitrary unverified sender.** Sender verification is universal.

### Is mobile-number OTP free?

**No, not for real users.** Trial credits and test numbers exist, but production SMS is paid and, for Indian numbers, adds DLT/TRAI compliance. Firebase Phone Auth is billed per SMS after a small daily free allowance and requires Blaze; Twilio Verify charges per successful verification plus channel fees and limits trial sends to verified recipients.

### Recommendation

**Use Resend with a verified custom domain** (for example `auth.onechat.app` or `mail.yourdomain.com`) and restore the existing Resend code path. This is the best fit because:

- the Go SDK is already in `go.mod` (`github.com/resend/resend-go/v3`);
- Resend is an HTTPS API, which works on Render free services;
- free quota is 100 emails/day and 3,000/month, enough for an early-stage app;
- no SMTP ports are involved;
- after domain verification, you can send to any recipient.

If you cannot get a domain today, use **Brevo free plan with a verified single sender as a short-term bridge**, but it is not a good permanent answer: free sender domains are replaced/flagged and Gmail/Yahoo/Microsoft require a custom authenticated domain. Do **not** switch to SMS for cost reasons.

### Other two errors

- `relation "channels" does not exist`: production has never run the `migrate` path. In `database.go`, migrations only run when the process is started with argument `migrate`, but the Dockerfile starts `air` and no `render.yaml` was found. Run migrations as a one-off command or start the service with the migration path enabled.
- `prepared statement name is already in use`: GORM's pgx driver enables prepared statement caching by default. The line `PreferSimpleProtocol: true` is already present but commented out in `database.go`; enabling it disables implicit prepared statements and is the appropriate fix here.

---

## 2. Evidence from the local backend

| Area | File | Finding |
|---|---|---|
| Registration | `internals/services/authService.go` | OTP generation/storage/send block is commented out; new user has no `VerifyOTPHash`, `VerifySalt`, or `OTPExpiresAt`. |
| OTP fields | `internals/db/models/user.go` | Fields already exist: `VerifyOTPHash`, `VerifySalt`, `OTPExpiresAt`, `Verified`. They are just not populated during `Register()`. |
| Resend flow | `internals/services/resendOTP.go` | Generates/hashes/stores OTP correctly, then calls `s.emailService.SendOTP()`. |
| Email client | `internals/services/emailService.go` | `NewEmailService()` returns an empty struct; `SendOTP()` calls `e.client.Emails.Send(...)` on a nil client. |
| Config | `internals/config/config.go` | Entire `Email` struct is commented out; no `RESEND_API`/`EMAIL_FROM`/`SEND_GRID_KEY` parsing. |
| Routes | `router/router.go` | `/auth/register`, `/auth/verify-otp`, `/auth/resend-otp` are registered. |
| DB connect | `internals/db/database.go` | `PreferSimpleProtocol: true` is commented; migrations only run with CLI arg `migrate`. |
| Deploy | `Dockerfile` | `CMD ["air"]`; no migration argument. No `render.yaml` in the repo. |
| Channels query | `internals/repository/channel_repo.go` | `FindForUser` joins `channel_members` to `channels`; query is correct once migrations have created the tables. |

The production log `POST /auth/register → 201` is therefore expected: the endpoint reports success immediately after creating the user, even though no email is sent.

---

## 3. Delivery method comparison

### 3.1 Render constraints

Render changed free web services so outbound SMTP to ports `25`, `465`, and `587` is blocked. Paid Render plans restore `465`/`587`; port `25` remains blocked on all plans because of the underlying EC2 environment. HTTPS API calls are not affected.

Sources:
- [Render changelog: free web services block SMTP ports](https://render.com/changelog/free-web-services-will-no-longer-allow-outbound-traffic-to-smtp-ports)
- [Render free service limits](https://render.com/docs/free)
- [Resend API introduction — HTTPS enforced](https://resend.com/docs/api-reference/introduction)

**Conclusion:** do not go back to raw SMTP while the service is on Render free. Use an HTTPS email API.

### 3.2 Email providers

| Option | Free allowance | Sender requirement | Sends to arbitrary recipients? | Fit for Onechat |
|---|---:|---|---|---|
| **Resend** | 100/day, 3,000/month free; 3 verified domains | Must verify a custom domain; `resend.dev` only sends to account owner | Yes, after domain verification | **Best choice.** SDK already present, simple API, HTTPS, cheap. |
| SendGrid | Free trial 100/day for 60 days (docs also describe 100/day free account) | Single Sender Verification works for testing; domain auth required for production/upgrades | Yes after sender verification | Workable fallback, but more account verification friction and deliverability caveats. |
| Brevo | 300/day free | Verify sender; custom authenticated domain strongly required | Yes, but free/public sender domains get replaced or filtered | Good fallback if you need no-card today, still needs a domain for reliable delivery. |
| SMTP (Gmail/custom) | Depends on provider | SMTP auth + verified sender | Yes | **Not on Render free** because SMTP ports are blocked. |
| Amazon SES | Free tier credits initially; then ~$0.16/1,000 emails in Essentials plan | Domain verification; production access request | Yes | Powerful but heavy setup for this stage. |

Sources:
- [Resend quotas](https://resend.com/docs/knowledge-base/account-quotas-and-limits)
- [Resend verified domains](https://resend.com/docs/dashboard/domains/introduction)
- [Resend 403 when using `resend.dev`](https://resend.com/docs/knowledge-base/403-error-resend-dev-domain)
- [SendGrid verifying your account](https://docs.sendgrid.com/ui/account-and-settings/verifying-your-account)
- [SendGrid single sender verification](https://docs.sendgrid.com/ui/sending-email/sender-verification)
- [SendGrid quickstart: 100/day trial and domain auth note](https://docs.sendgrid.com/for-developers/sending-email/quickstart-nodejs)
- [Brevo free plan limits](https://help.brevo.com/hc/en-us/articles/208580669-FAQs-What-are-the-limits-of-the-Free-plan)
- [Brevo sender/domain requirements](https://help.brevo.com/hc/en-us/articles/14925263522578-Comply-with-Gmail-Yahoo-and-Microsoft-s-requirements-for-email-senders)
- [Amazon SES pricing](https://aws.amazon.com/ses/pricing/)

### 3.3 Why your Resend test only worked for one email

The exact Resend error is documented:

> “You can only send testing emails to your own email address … To send emails to other recipients, please verify a domain.”

So the fix is **not** to change providers; it is to verify a domain and set `EMAIL_FROM` to an address on that domain. After verification, you do not need to “create” the sender mailbox in Resend.

### 3.4 SMS OTP cost and complexity

**Firebase Phone Auth**
- Spark (free) plan: Phone Auth is listed as “Not applicable”.
- Blaze plan: billed per SMS. The first 10 SMS per day are not billed; after that, India is `$0.07` per SMS in the current Identity Platform pricing table.
- Web requires reCAPTCHA, an SMS region policy, and authorised domains.
- Test fictional phone numbers do not consume quota or send SMS, but they are for development only.

Sources:
- [Firebase pricing](https://firebase.google.com/pricing)
- [Google Identity Platform pricing](https://cloud.google.com/identity-platform/pricing)
- [Firebase phone auth web guide](https://firebase.google.com/docs/auth/web/phone-auth)

**Twilio Verify**
- Trial is free for 30 days but can only send to verified recipient numbers.
- Production pricing starts at `$0.05` per successful verification plus channel fees (for example US SMS `$0.0083` per SMS in the published table).
- Sending to any recipient requires an upgraded account and an approved Primary Compliance Profile.

Sources:
- [Twilio Verify pricing](https://www.twilio.com/en-us/verify/pricing)
- [Twilio Verify quickstarts / trial restrictions](https://www.twilio.com/docs/verify/quickstarts)

**India-specific note**
TRAI requires entities sending OTP/transactional/service SMS to register as a Principal Entity, register a header, and register content templates on DLT. An OTP message must match the registered template. This is additional setup even if you pay for SMS.

Source: [TRAI — Advice to Senders](https://trai.gov.in/advice-to-senders)

**Conclusion:** SMS is not free, adds phone-number collection/UI changes, and adds compliance overhead. It is not the right next move for Onechat.

### 3.5 A genuinely free alternative: TOTP authenticator app

TOTP codes generated locally by Google Authenticator/Authy-style apps cost nothing per verification and work offline. Twilio and Firebase both support TOTP products. However, this changes the user flow: the user must install an app and scan/store a seed. It is a good later option for 2FA, not a drop-in replacement for email OTP during signup.

Source: [Twilio Verify TOTP overview](https://www.twilio.com/docs/verify/quickstarts/totp)

---

### 3.6 How Telegram does login codes — and Telegram Gateway

Telegram is a special case:

- It owns the app and the delivery network. For its own login it can deliver a code **inside Telegram** to an already logged-in device, by **SMS**, or by **voice call**. Passkeys can avoid SMS entirely.
- First-time users still need a phone number. Telegram does not have a free magic route that bypasses carriers for users who cannot receive anything in the Telegram app.
- When Telegram uses SMS, it pays carriers/aggregators like everyone else, but at enormous scale with direct/local routes and much lower unit costs than a small app buying Twilio/Firebase messages.
- App-delivered codes use Telegram's own internet infrastructure, so there is no per-SMS carrier charge for that path.

Telegram also sells this as **Telegram Gateway** for third-party apps: official pricing is **$0.01 per delivered code**, with a free testing environment (sending to your own Telegram number is free). Users must have a Telegram account, share their phone number voluntarily, and opt in to receive messages. It is an HTTPS API, so it works from Render.

This is an interesting secondary option, but not a replacement for email OTP in Onechat:

- users without Telegram are excluded;
- you must collect phone numbers and handle opt-in/privacy;
- it changes the registration UI and data model;
- Telegram's terms note availability can vary by region/regulation, and credit purchases are prepaid/non-withdrawable.

Sources:
- [Telegram FAQ — login/SMS/call/in-app codes and passkeys](https://telegram.org/faq)
- [Telegram Gateway overview](https://core.telegram.org/gateway)
- [Telegram Gateway API](https://core.telegram.org/gateway/api)
- [Telegram Gateway terms](https://telegram.org/tos/gateway)

---

## 4. Recommended implementation plan — Resend email OTP

### 4.1 What you need to do once outside the code

1. Buy/control a domain if you do not already have one.
2. In Resend, add and verify the domain (or a subdomain such as `auth.yourdomain.com`). Add the DNS records Resend gives you.
3. Create a Resend API key with send permission.
4. In Render, set environment variables:
   - `EMAIL_PROVIDER=resend` (optional but useful)
   - `RESEND_API=<your Resend API key>`
   - `EMAIL_FROM=Onechat <auth@yourdomain.com>`
5. Do **not** set `onboarding@resend.dev` as `EMAIL_FROM` in production.

### 4.2 Backend code changes to make after you approve

1. **Restore `config.AppConfig.Email`** in `internals/config/config.go`:
   - `Provider`, `From`, `ResendAPI`, `SendGridKey`.
   - Avoid `required` tags that force SendGrid when only Resend is configured. Validate the selected provider explicitly instead.
2. **Fix `NewEmailService()`** in `internals/services/emailService.go`:
   - Pass config into it from `app.go`.
   - Initialise `resend.NewClient(cfg.Email.ResendAPI)`.
   - Set `from` from `EMAIL_FROM`.
   - Fail loudly at startup if the chosen provider is missing its key, instead of panicking on first registration.
3. **Restore OTP in `AuthService.Register()`**:
   - `utils.GenerateOtp()`
   - `salt := uuid.NewString()`
   - `utils.HashOTP(otp, salt)`
   - `expiresAt := time.Now().Add(5 * time.Minute)`
   - Populate `VerifyOTPHash`, `VerifySalt`, `OTPExpiresAt` on the new user.
   - Create the user, then send the email.
   - If the send fails, either return an error and clean up the created user, or keep the user but make the response explicit so the UI can offer “Resend OTP”. Do not silently return `201` with no delivery, which is the current bug.
4. **Remove the misleading `SendOTPWITHSMTP` method** or rename it to what it actually does (Resend). Keep `SendOTPWithGrid` only if you really want SendGrid fallback.
5. **Do not change the frontend OTP flow.** `otpApi.ts`, `registerApi.ts`, `useRegister.ts`, and `VerifyOTPForm.tsx` already expect email + OTP.

### 4.3 Tests/verification after implementation

- `go test ./...`
- Local smoke:
  - `POST /auth/register` creates user with OTP fields set.
  - `POST /auth/verify-otp` verifies correct code and rejects wrong/expired codes.
  - `POST /auth/resend-otp` generates a new code and sends email successfully.
- With a verified Resend domain, confirm delivery to a non-account email address (not only your own).
- Check Render logs for `📨 SendOTP called`, Resend errors, and no nil-client panic.

---

## 5. Production DB repair plan

### 5.1 Missing `channels` table

Root cause: migrations are not executed by the normal Render start command.

- `internals/db/database.go` only migrates when `os.Args[1] == "migrate"`.
- `Dockerfile` runs `CMD ["air"]`.
- No `render.yaml` was found in the backend repo.

**Immediate repair:** use Render Shell / a one-off job against the same `DATABASE_URL`:

```bash
go run . migrate
```

This runs `migrateChannels()` (legacy `crews` → `channels`, `crew_members` → `channel_members`, column renames), then `AutoMigrate` for all models, then the cascade/status migrations. The helper is written to be repeat-safe, but take a DB backup/snapshot before running it against production.

**Durable fix options:**

- Preferred: keep migrations explicit and add a Render one-off job/`pre-deploy` command that runs `go run . migrate` before the web service restarts.
- Alternative: re-enable the already-commented `RUN_MIGRATIONS` gate and set `RUN_MIGRATIONS=true` on Render, so startup runs migrations automatically. Only do this if you accept migrations on every restart.

Once tables exist, `GET /api/channels/` should stop failing with `42P01`.

### 5.2 Prepared statement name already in use

Root cause: GORM's PostgreSQL driver uses pgx with implicit prepared statements by default. After a failed query or across pooled/reused connections, the same prepared statement name can be reused unexpectedly (`08P01`).

**Fix:** in `internals/db/database.go`, enable the already-present config line:

```go
DB, err = gorm.Open(postgres.New(postgres.Config{
    DSN:                  dbe,
    PreferSimpleProtocol: true,
}), &gorm.Config{})
```

GORM documents `PreferSimpleProtocol: true` as disabling implicit prepared statement usage. Source: [GORM — Connecting to a Database](https://gorm.io/docs/connecting_to_the_database.html).

Optional hardening if it persists: cap `SetMaxOpenConns` and set `SetConnMaxLifetime` on the underlying `sql.DB`, but do that only if needed.

**Do not change `channel_repo.go` yet.** Its join is valid once the table exists.

---

## 6. Decision needed before code changes

My recommendation is: **Resend + verified custom domain**.

Please confirm one of these paths:

1. **You have/control a domain** — I restore the Resend OTP path, you add `RESEND_API` and `EMAIL_FROM` on Render after verifying the domain.
2. **You do not have a domain yet** — either get one (best long-term), or I can wire Brevo/SendGrid single-sender as a temporary bridge, accepting deliverability/verification caveats.
3. **You still want SMS** — not recommended; it requires a paid/trial provider, phone-number UI changes, and Indian DLT compliance.

After that approval, I can implement the OTP fixes and the two DB fixes together, then run `go test ./...` and a local registration/verification smoke test.

---

## 7. Source list

All provider/pricing claims above were checked on 2026-11-24 against these primary sources:

- Render SMTP changelog: https://render.com/changelog/free-web-services-will-no-longer-allow-outbound-traffic-to-smtp-ports
- Render free limits: https://render.com/docs/free
- Resend quotas: https://resend.com/docs/knowledge-base/account-quotas-and-limits
- Resend domains: https://resend.com/docs/dashboard/domains/introduction
- Resend `resend.dev` restriction: https://resend.com/docs/knowledge-base/403-error-resend-dev-domain
- Resend API HTTPS: https://resend.com/docs/api-reference/introduction
- SendGrid verification: https://docs.sendgrid.com/ui/account-and-settings/verifying-your-account
- SendGrid single sender: https://docs.sendgrid.com/ui/sending-email/sender-verification
- SendGrid quickstart: https://docs.sendgrid.com/for-developers/sending-email/quickstart-nodejs
- Brevo free limits: https://help.brevo.com/hc/en-us/articles/208580669-FAQs-What-are-the-limits-of-the-Free-plan
- Brevo sender requirements: https://help.brevo.com/hc/en-us/articles/14925263522578-Comply-with-Gmail-Yahoo-and-Microsoft-s-requirements-for-email-senders
- Firebase pricing: https://firebase.google.com/pricing
- Google Identity Platform pricing: https://cloud.google.com/identity-platform/pricing
- Firebase phone auth: https://firebase.google.com/docs/auth/web/phone-auth
- Twilio Verify pricing: https://www.twilio.com/en-us/verify/pricing
- Twilio Verify restrictions: https://www.twilio.com/docs/verify/quickstarts
- TRAI sender requirements: https://trai.gov.in/advice-to-senders
- GORM PostgreSQL config: https://gorm.io/docs/connecting_to_the_database.html
- Amazon SES pricing: https://aws.amazon.com/ses/pricing/
