# helm-rename



## Overview

The plugin intends to rename existing release from a previous name to a new name, allowing upgrades on the new name.

The mechanism of the plugin:
- creating helm releases (secrets) in the new name, based on the previous - and removing the old secrets.
- annotating the resources with owenership annotaion - so upgrade won't fail.

This doesn't ensure a successful upgrade post-rename! Please read the next section about things to be cautious about.

## CAUTION: Readme before rename
Things to consider:
- If your resources names are based on the release name, renaming the release and then upgrading without changing the chart will lead to deletion of objects, and creation of new ones. if fullNameOverride (if you use it) isn't set and you use it - set it to the previous name.
- Standart chart (`helm create foo`) set selectorLabels with release name, with no option to override it for users. As selectorLabels in k8s are currently immutable, this can lead to failure in helm upgrade (post-upgrade)
- Anyway, the first helm upgrade post rename is very dangerous. 

## When you should use this plugin:
Renaming a release is very dangerous, yet very useful. As said before, this is very dangrous and probably the best way to do it is reinstalling the release.

- If you want to rename a release you know (or think) won't ever be upgraded with helm anymore - it's pretty safe. This can help for legacy applications, or an applcation that is being replaced by different application. As the operation is revertable, you can always go back to the previous name.
- When you know that Release.Name isn't used anywhere
 

 ## Other

 I don't really write in Golang, so:
 - Be careful using it
 - if you happen to open the code and see something that I did like an idiot (or anything that can be done better) - please write to me/open PR/open an issue.
 - Thanks - Many of the code is inspired by helm, helm-2to-3 plugin and helm-edit plugin (some of the I used as a "base" code that I changed, sometimes with no relation to the resulted method)

 I will be glad to get any PRs/Issues/Comments