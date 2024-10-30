# magiclinksdev

The **magiclinksdev** project is an authentication service for magic link and One-Time Password (OTP) use cases. There
is built-in email support through Amazon SES and SendGrid.

Use cases include:
* Sign up
* Log in
* Password resets
* Email verification
* And more authentication use cases

This project can be used to supplement password authentication or replace it entirely.

If your project has an alternate secure means of communication, you can use generate magic links and OTPs without
sending emails. An example would be mobile push notifications.

## Getting started

To get started implementing a client application that uses **magiclinksdev** for authentication, the recommended path
is:
1. Do the [quickstart](https://docs.magiclinks.dev/self-host-quickstart)
2. Find a [pre-built SDK](https://docs.magiclinks.dev/client-sdk) or [generate one from the formatted API specification](https://docs.magiclinks.dev/client-api-specification#generate-code)
3. Choose the [magic link](https://docs.magiclinks.dev/client-magic-link-workflow) or [OTP](https://docs.magiclinks.dev/client-otp-workflow) workflow
4. Review the [implementation tips](https://docs.magiclinks.dev/client-implementation-tips) for recommendations and best practices

## Screenshots

The built-in email templates are friendly to mobile and desktop screens. They also adapt to light/dark mode
automatically. The templates are built using [maizzle](https://maizzle.com/).

<span>
  <img width="400" src="https://magiclinks.dev/screenshots/magic-link-email-light-mobile-example.png" alt=""/>
  <img width="400" src="https://magiclinks.dev/screenshots/magic-link-email-dark-mobile-example.png" alt=""/>
</span>
<span>
  <img width="400" src="https://magiclinks.dev/screenshots/otp-email-light-mobile-example.png" alt=""/>
  <img width="400" src="https://magiclinks.dev/screenshots/otp-email-dark-mobile-example.png" alt=""/>
</span>

## Suggested Magic Link Workflow
<img width="1000" src="https://magiclinks.dev/mermaid/suggested-magic-link-workflow.png" alt=""/>

## Suggested OTP Workflow
<img width="1000" src="https://magiclinks.dev/mermaid/suggested-otp-workflow.png" alt=""/>

## Self-hosting the service

The **magiclinksdev** project is open-source and can be self-hosted. Check out the [**Quickstart**](https://docs.magiclinks.dev/self-host-quickstart) page
to get started in minutes. For reference on configuring your self-hosted instance, check out the
[**Configuration**](https://docs.magiclinks.dev/self-host-configuration).

## Source code and license

The **magiclinksdev** project is [open source on GitHub](https://github.com/MicahParks/magiclinksdev) and licensed
under [**Apache License 2.0**](https://github.com/MicahParks/magiclinksdev/blob/master/LICENSE).

## Optional SaaS platform

You can find the optional Software-as-a-Service (SaaS) platform landing page at https://magiclinks.dev. Use of the SaaS
platform is not required, but it's very inexpensive and may be cheaper than deploying yourself.

## Support the project

This project took a lot of time, effort, and money to create and maintain for you. If you get some business value of
this project consider becoming a [GitHub Sponsor](https://github.com/sponsors/MicahParks).
