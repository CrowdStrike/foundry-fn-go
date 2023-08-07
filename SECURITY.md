![CrowdStrike Falcon](/docs/asset/cs-logo.png?raw=true)

# Security Policy

This document outlines security policy and procedures for the CrowdStrike `foundry-fn-go` project.

+ [Supported Go versions](#supported-go-versions)
+ [Supported CrowdStrike regions](#supported-crowdstrike-regions)
+ [Supported foundry-fn-go versions](#supported-foundry-fn-go-versions)
+ [Reporting a potential security vulnerability](#reporting-a-potential-security-vulnerability)
+ [Disclosure and Mitigation Process](#disclosure-and-mitigation-process)

## Supported Go versions

foundry-fn-go functionality is unit tested to run under the following versions of Go. Unit testing is performed with
every pull request or commit to `main`.

| Version  |                    Supported                    |
|:---------|:-----------------------------------------------:|
| \>= 1.19 | ![Yes](https://img.shields.io/badge/-YES-green) |
| <= 1.18  |   ![No](https://img.shields.io/badge/-NO-red)   |


## Supported CrowdStrike regions

foundry-fn-go is unit tested for functionality across all non-gov CrowdStrike regions.

| Region | 
|:-------|
| US-1   |
| US-2   |
| EU-1   |

## Supported foundry-fn-go versions

When discovered, we release security vulnerability patches for the most recent release at an accelerated cadence.

## Reporting a potential security vulnerability

We have multiple avenues to receive security-related vulnerability reports.

Please report suspected security vulnerabilities by:

+ Submitting
  a [bug](https://github.com/CrowdStrike/foundry-fn-go/issues/new?assignees=&labels=bug+%3Abug%3A&template=bug_report.md&title=%5B+BUG+%5D+...).
+ Starting a new [discussion](https://github.com/CrowdStrike/foundry-fn-go/discussions).
+ Submitting a [pull request](https://github.com/CrowdStrike/foundry-fn-go/pulls) to potentially resolve the issue. (New
  contributors: please review the content
  located [here](https://github.com/CrowdStrike/foundry-fn-go/blob/main/CONTRIBUTING.md).)
+ Sending an email to __foundry-fn-go@crowdstrike.com__.

## Disclosure and mitigation process

Upon receiving a security bug report, the issue will be assigned to one of the project maintainers. This person will
coordinate the related fix and release process, involving the following steps:

+ Communicate with you to confirm we have received the report and provide you with a status update.
    - You should receive this message within 48 - 72 business hours.
+ Confirmation of the issue and a determination of affected versions.
+ An audit of the codebase to find any potentially similar problems.
+ Preparation of patches for all releases still under maintenance.
    - These patches will be submitted as a separate pull request and contain a version update.
    - This pull request will be flagged as a security fix.
    - Once merged, and after post-merge unit testing has been completed, the patch will be immediately published.

---

<p align="center"><img src="https://raw.githubusercontent.com/CrowdStrike/falconpy/main/docs/asset/cs-logo-footer.png"><BR/><img width="300px" src="https://raw.githubusercontent.com/CrowdStrike/falconpy/main/docs/asset/adversary-goblin-panda.png"></P>
<h3><P align="center">WE STOP BREACHES</P></h3>