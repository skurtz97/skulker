# Skulker

Based on the name of a program I encountered in z/OS UNIX, and an attempt to recreate it for sentimental reasons.

## Background

On IBM's enterprise operating systems (currently IBM Z but previously known by many names such as OS/360, OS/370, MVS, etc.) that run currently on the s390x instruction set — and in many of the older original variants of programs — you really did need a cron job / JCL job to run Skulker (or something much like it) every once in a while. Clean-up of certain things was not built-in to the operating system and it was expected that the systems programmers (and applications developers if their programs left anything laying around in strange places) would program their own cleanup scripts, because every site was wildly different in how hey setup their operating systems (the existance of exits for example are something that remind of a little bit of eBPF in Linux). It would be a fair analogy to say that every mainframe using organization had their own "distro" in modern parlance.

## What It Does

Skulker is mostly a recreation of that logic with a friendlier TUI, written in Go with the help of . It:

1. Deletes files not touched in a configurable period of time
2. Cleans up stale `/tmp` files
3. Removes duplicates
4. And more

## Motivation

In other words, you used to have to program your computer to do things like clean up after itself. Now it does it for you — but it always did; it just required a system programmer to go write a cron job (or JCL job or possibly REXX, if you were in the IBM world). There is something nostalgic about building one for yourself in a language you like (Go).

