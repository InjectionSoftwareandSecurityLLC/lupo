# Contributors Guide

Thanks for taking interest in Lupo C2! 

If you're here that means you are likely someone who is already contributing to Lupo C2 to make it better, or you are interested in learning how to contribute to Lupo C2.

If you'd like to contribute monetarily consider donating to this project via the Github Donate button!

Either way welcome to the project. This guide will walk you through the basics of submitting contributions to Lupo C2.

Contributing to Lupo C2 is very easy. You just need to follow a few simple guidelines to insure your code is submitted correctly. *Code that is submitted incorrectly might have to be turned away in order to maintain the integrity of the repository.* Keep that in mind, because even we would hate to turn away contributions to the project especially if they are amazing and just submitted incorrectly.

If you have never contributed to an open source project on Github before, then we suggest you read this guide first:

[How to Contribute](https://opensource.guide/how-to-contribute/)

We also use the `git flow` methodology, so if you're not familiar with that please read up here before contributing:
[Git Flow](https://guides.github.com/introduction/flow/)

Below is a step by step guide on how to create a contribution with git and submit a Pull Request for it to be merged into the project.

## Contribution Git Workflow:

If you are not a maintainer of Lupo C2, you will first need to fork the repo, and then follow this workflow. If you are a maintainer it is not necessary to fork the repo, but you can do whatever makes you happy (you are a maintainer after all).


#### 1. Getting Started:

When creating something for Lupo C2, start by creating a new branch in Git.

`git branch dev/my-awesome-new-feature`

Then checkout your new branch...

`git checkout dev/my-awesome-new-feature`

Now start writing that awesome new feature!

#### 2. Pushing Up:

Once you're done making your awesome new feature, simply push it up!

`git push origin dev/my-awesome-new-feature`

That's it. Next the step is about keeping your feature up to date.

#### 3. Staying Updated (Rebasing):

As you build your new feature the main project may get updated. Especially if everyone's building the latest and greatest modifications to Lupo C2, the project could update very frequently!

In the event that this happens, you will need to rebase your branch to stay up to date.

Rebasing allows you to resolve merge conflicts on your end, and submit a clean, up to date branch in a pull-request.

Merge conflicts are bound to happen eventually, but if every contributor is keeping up with the repo and utilizing git flow, we should hopefully avoid most problems when it comes time for the final pull-request.

To perform a rebase. Make sure you are on your feature's branch:

`git checkout dev/my-awesome-new-feature`

Then simply rebase!

`git rebase origin/main`

This will rebase your current branch onto the latest version of the master branch.

**Always rebase before you push up or submit a pull request!**


#### 3. Make a Pull Request!

Now your super aweomse feature is done. It's time to make a request to the Lupo C2 project maintainers to merge in your feature.

To do so, follow these steps.

1. Log into Github
2. Visit the Lupo C2 repo
3. Click on the button that says "Compare & Pull Request"
4. Fill out the Pull Request.

The Following image shows the main parts of the Pull Request that need to by populated before submission. You will need to name your request, and add in a description of the feature in the appropriate text boxes provided during the request creation process.

The three sections that are outlined in the image below are required in order to submit your request. 
    
Especially make sure you specifiy your feature type in the "Labels" field. This will help us prioritize Pull Requests a little better since we will be able to differintiate between whether a feature is a bug fix, or a new feature based on the label you provide.

5. Submit the Pull Request!

And now you wait... :)

A maintainer will review your submission as soon as possible. Depending on your feature we may have to do some rigourous testing to make sure it won't break anything, so please be patient. If we run into any issues we will communicate with you to resolve them and do our best to get your contribution merged in.

### That's it, thanks for contributing to Lupo C2!
