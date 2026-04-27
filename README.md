# Skulker

Based on the name of a program I encountered in z/OS UNIX, and an attempt to recreate it for sentimental reasons. Currently, only intended for use on Fedora or other RedHat derived linux distros.

## Background

On IBM's enterprise operating systems (currently z/OS but previously known by many names such as OS/360, OS/370, MVS, etc. going well back to before Unix was around and popular) that run currently on the s390x instruction set — and in many of the older original variants of programs — you really did need a cron job / JCL job to run Skulker (or something much like it) every once in a while. Clean-up of certain things was not built-in to the operating system and it was expected that the systems programmers (and applications developers if their programs left anything laying around in strange places) would program their own cleanup scripts, because every site was wildly different in how hey setup their operating systems (the existance of exits for example are something that remind of a little bit of eBPF in Linux). It would be a fair analogy to say that every mainframe using organization had their own "distro" in modern parlance.

## What It Does

Skulker is mostly a recreation of that logic with a friendlier TUI, written in Go with the help of BubbleTea.

1. Deletes DNF metadata and cache.
2. Runs `dnf autoremove`.
3. Removes unused flatpak runtimes.
5. Cleans up the systemd journal.
6. Removes old kernels.
7. Purges the user cache.

And more as I have more time to work on it and find more areas to safely cleanup. It's also a good excuse to play around with [BubbleTea](https://github.com/charmbracelet/bubbletea), a TUI framework that I think looks marvelous.

<img width="741" height="279" alt="image" src="https://github.com/user-attachments/assets/7b7d3802-5a92-4497-850e-2b15cd6b5ad3" />
<br/>
<img width="364" height="325" alt="image" src="https://github.com/user-attachments/assets/ac76ccc3-838d-4da6-9ee5-113b661890e1" />


## Motivation

In other words, you used to have to program your computer to do things like clean up after itself and you'd have a particular group of people that would be responsible for system stability and cleanup like that. Computers used to be HARD believe it or not. Fedora and RHEL currently clean up after themselves just fine for ordinary use and so this isn't a necessary program or anything, just an optional convenience you can use. Clearing caches may also be useful for debugging in certain instances. Just don't go running around using it every 5 seconds...things like the journal and cache are there for a good reason, they just sometimes become bugged or bloated. For instance, after a system snapshot would be a good time to run some cleanup.
