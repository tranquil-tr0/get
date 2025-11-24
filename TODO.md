- support prereleases
- add install version ability in GUI, consider reworking install ux in gui to use an install modal
- support github release links for specific version install
- support direct links to autoupdating files like beeper and discord
- replace tag name prefix filtering with tag name regex filtering instead
- support tag name filtering in gui
- improve same file recognition when file name contains iterating version
- make upgrade command in cli accept package names afterward (get upgrade tranquil-tr0/get jj-vcs/jj)
- dont log a package as installed in json until actually installed, same issue appears to exist in the upgrade code as well
- stop using pkexec
- all the // TODO: <task>
- stop the gui from becoming frozen/unresponsive when executing task, such as by showing loading spinner instead
- improve the look of things overall
- support rpm
- improve the cli list to look better and show package type
- fix gui not resizable to be narrower
- fix gui icon not in (my) taskbar (probably need a smaller icon) (for the deb)

IMPORTANT dont make a json file in the code, it should be created as part of the deb. (or install script if I make one later)
