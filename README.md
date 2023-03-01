# magiclinksdev

**This project has not been released yet.**

You can find the documentation for this project on the [docs site](https://docs.magiclinks.dev). This site contains
resources for implementing a client and self-hosting the project.

You can find the SaaS landing page at [https://magiclinks.dev](https://magiclinks.dev). Use of the SaaS platform is not
required, but it's very inexpensive and may be cheaper than deploying yourself.

# Getting started

The **magiclinksdev** project is an authentication service that uses magic links to authenticate users. A typical use
case would involve sending a magic link to a user via email. After the user clicks the link, a new authenticated session
is created for that user. Sometimes **magiclinksdev** is abbreviated as "mld".

## About

This project is a magic link authentication service. It serves use cases like:

* Sign up
* Log in
* Password resets
* Email verification
* And more authentication use cases

It can be used to supplement password authentication or replace it entirely.

A typical use case involves sending a magic link to a user via email. After the user clicks the link, a new
authenticated session is created for that user.

## Screenshots

The built-in email template is populated on a per-request basis. It adapts to the device's theme automatically. This
template was built using [maizzle](https://maizzle.com/).

<span>
    <img width="400" src="https://magiclinks.dev/screenshots/mobile-light.png" alt="">
    <img width="400" src="https://magiclinks.dev/screenshots/mobile-dark.png" alt="">
</span>

## Suggested Email Workflow

<img width="1000" src="https://magiclinks.dev/mermaid/suggested-email-workflow.png" alt=""/>

## Implementing a client application

Client applications are programs that use the **magiclinksdev** project to authenticate their users. Check out the
[**SDKs**](https://docs.magiclinks.dev/sdks) page to get started with an existing SDK. If you can't find an SDK for your
language, you can use the [**Specification**](https://docs.magiclinks.dev/specification) to implement your own client by
hand or code generation. To learn more about the client workflow, check out the
[**Workflow**](https://docs.magiclinks.dev/workflow) page.

## Self-hosting the service

The **magiclinksdev** project can be self-hosted. Check out the [**Quickstart**](https://docs.magiclinks.dev/quickstart)
page to get started in minutes. For reference on configuring your self-hosted instance, check out the
[**Configuration**](https://docs.magiclinks.dev/configuration).

## Source code and license

The **magiclinksdev** project is [publicly available on GitHub](https://github.com/MicahParks/magiclinksdev) and
licensed
under [**Elastic License 2.0 (ELv2)**](https://github.com/MicahParks/magiclinksdev/blob/master/LICENSE). The ELv2 is a
permissive license with three simple limitations. Please see
the [ELv2 FAQ](https://www.elastic.co/licensing/elastic-license/faq) for more information.

## Optional SaaS platform

You can find the optional Software-as-a-Service (SaaS) platform landing page at https://magiclinks.dev. Use of the SaaS
platform is not required, but it's very inexpensive and may be cheaper than deploying yourself.

## Support the project

This project took a lot of time, effort, and money to create and maintain for you. If you get some business value of
this project consider becoming a [GitHub Sponsor](https://github.com/sponsors/MicahParks).
