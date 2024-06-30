# Contributing to Photoview

👋 Welcome to Photoview! Thank you for considering contributing to our project. Before you start, 
please take a moment to review this guide, which outlines the process for making contributions.

## Contents

- [Getting Started](#getting-started)
- [Reporting Issues](#reporting-issues)
- [Making Contributions](#making-contributions)
  - [Here's how you can contribute](#heres-how-you-can-contribute)
  - [Ticket workflows](#ticket-workflows)
    - [Bug](#-bug-)
    - [Feature](#-feature-)
  - [Steps to contribute to the code or documentation](#steps-to-contribute-to-the-code-or-documentation)
- [Code Standards](#code-standards)
- [Documentation](#documentation)
- [Licensing](#licensing)
- [How to get help](#how-to-get-help)

## Getting Started

1. **Join project's Discord server**: We have the [Discord server](https://discord.gg/jQ392948u9) 
with several channels, and you're welcome to join it to chat with the community and maintainers 
to ask for help, discuss some project-related topics, or help others.

2. **Browse Issues**: Check the [open issues](https://github.com/photoview/photoview/issues), 
[PRs](https://github.com/photoview/photoview/pulls), and [discussions](https://github.com/photoview/photoview/discussions)  
to see if your bug or idea has already been reported or discussed. If not, you can open a new issue.
While browsing the issues, if you see an issue with useful from your PoV feature proposal or a bug, 
you're struggling with, it would be a helpful contribution to like the 1st post of that issue in the 
conversation tab (the issue description) - in this way, maintainers will know which issues should take priority.
Of course, you can also add a comment with some useful info, which might be missing in the 
issue's thread - that would help as well.

3. **Check Open Source Guide**: Familiarize yourself with the [Open Source Guide](https://opensource.guide/) 
for general information on contributing to open-source projects.

4. **Review Good First Issues**: We've collected some [good first issues](https://github.com/photoview/photoview/labels/good%20first%20issue) 
that are suitable for newcomers. These are a great way to start contributing to Photoview.

## Reporting Issues

If you encounter a bug, have a suggestion, or want to discuss an idea, please open a new issue or start a discussion. 
When opening an issue or starting a discussion, provide detailed information, including steps to reproduce the problem 
or a clear description of your idea.

## Making Contributions

We welcome contributions in various forms, including bug fixes, feature enhancements, 
documentation improvements, and more. If you have the skills, it doesn't matter how small the contribution is welcome!

### Here's how you can contribute

- **Code**: If you're looking to help with code, check out the open issues. This 
[filter](https://github.com/photoview/photoview/issues?q=is%3Aopen+is%3Aissue+project%3Aphotoview%2Fphotoview%2F1+-label%3Aduplicate+-label%3Ainvalid+-label%3Awontfix+no%3Aassignee)
contains issues, reviewed by maintainers, and waiting for contribution. 
In particular, they have the Photoview project assigned and don't have any invalid labels.
- **Code review**: If you're a skilled developer with experience in the project's technology stack,
we'd appreciate your help with code reviews of open PRs, providing well-described expert-level feedback 
with all info, needed for a contributor of any level of seniority to fix or optimize the PR.
- **Documentation**: Spotted a typo, or think a document needs clarification? 
Go ahead and suggest changes in the [documentation repo](https://github.com/photoview/photoview.github.io).
- **Ideas**: Have an idea for a new feature? We'd love to hear it!
Submit a new feature request issue.

### Ticket workflows

#### < Bug >

The bug flow is the simplest of the two flows, and will most likely jump straight to development unless 
further discussion on architectural changes is required.

Development -> Code review -> Merge

#### < Feature >

New features are intended to be the longest flow as we want to ensure there is no time wasted 
by contributors. And that anyone can contribute to Photoview.

Discussion -> Dev Approach -> Development -> Code Review -> Merge

> NOTE: Some features if basic may be able to skip stages.

**Discussion**

A new feature will start its life as an issue tagged as such, the initial author should highlight 
what they want from the feature. Then from this, interested community members can spark discussion 
and ideas about how this should work from a user flow perspective, for instance:

- User navigates to settings
- There is a specific section for `x`
- There will be an `Add` button next to the title

**Development approach**

When sufficient conversation has happened, a label will be added to the ticket indicating it is ready 
to be picked up for a development approach, this gives someone the ability to investigate the existing 
architecture and design a high-level solution overview, doing so ensures that work is implemented 
into the system in a maintainable way, and also, hopefully, reduces the amount of rework when a code review 
happens. The development doesn't have to be done by the same person, and, hopefully, this will give 
less experienced members a chance to contribute to development.

**Development**

When the development approach has been written and signed off, it will be the time to develop the work. 
Again, this doesn't necessarily need to be done by the same person.

**Code review**

As with all PRs, the code will be reviewed by the community and maintainers before being merged ensuring 
the change meets the requirements, dev approach, and integrates correctly into the system. 
This will, hopefully, drive the community spirit and help people get support if they need it rather than giving up.

### Steps to contribute to the code or documentation

1. **Fork the Repository**: Fork the Photoview repository to your GitHub account.

2. **Clone the Repository**: Clone your forked repository to your local machine.

3. **Create a Branch**: Before making any changes, create a new branch to work on your contribution.

4. **Make Changes**: Implement your changes, ensuring they align with the project's standards and guidelines.

5. **Test Your Changes**: Before committing your changes, test them locally to verify their correctness 
and ensure they don't introduce any regressions.

6. **Commit Changes**: Once you're satisfied with your changes, commit them with clear and descriptive commit messages.

7. **Push Changes to Your Fork**: Push your changes to your forked repository on GitHub.

8. **Create a Pull Request**: [Create a pull request](https://github.com/photoview/photoview/compare) 
from your branch to the `master` branch of the Photoview repository. Provide a detailed description 
of your changes and reference any related issues.

9. **Review and Collaborate**: Collaborate with project maintainers and address any feedback or review comments on your pull request.

## Code Standards

When contributing code, please adhere to the following standards:

- Follow the existing coding style and conventions used throughout the project.
- Write clean, readable, and well-commented code.
- Ensure your changes pass all existing tests and write additional tests for new functionality.
- Ensure that existing Photoview users can migrate to the new version, containing your changes, 
as smoothly as possible: with no (or minimal) manual actions (which are well documented and clearly indicated to the user)
and no data loss.
- GitHub Actions, executed for your PR, shouldn't report issues, new to the `master` branch.
- Ensure your changes are documented in the project documentation, Readme, and other applicable places 
in the case, they change the existing user experience.

## Documentation

Improving documentation is a valuable contribution. If you find areas where documentation can be enhanced 
or if you have insights into better explaining existing features, feel free to update the relevant 
documentation files by creating a PR in the [documentation repo](https://github.com/photoview/photoview.github.io).

## Licensing

By contributing to Photoview, you agree that your contributions will be licensed under the terms of 
the [GNU Affero General Public License (AGPL) version 3](./LICENSE.txt). Ensure that your contributions comply with this license.

## How to get help

There are several ways to get help for your contribution to the project:

- Search for the answer in the project's [documentation](https://github.com/photoview/photoview.github.io), 
[issues](https://github.com/photoview/photoview/issues), [PRs](https://github.com/photoview/photoview/pulls), 
and [discussions](https://github.com/photoview/photoview/discussions).
- We have the [Discord server](https://discord.gg/jQ392948u9) with several channels, 
and you're welcome to join it to chat with the community and maintainers.
- Start a Discussion in the repo, providing a detailed and complete description of your case, environment, and the problem.