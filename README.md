# Skulker
 
Based on the name of a program I encountered in z/OS UNIX, and an attempt to recreate it for sentimental reasons.
 
## Background
 
On IBM's older-architecture operating systems for UNIX — and in many of the older original variants of programs — you really did need a cron job to run Skulker every once in a while. (New application development for the platform is strong and modernization is happening at a rapid pace; customers know it is a priority.) You could change the settings to leave certain things alone, like files in a RAM-based `/tmp` storage where often-running application files were kept.
 
## What It Does
 
Skulker is mostly a recreation of that logic with a friendlier TUI. It:
 
1. Deletes files not touched in a configurable period of time
2. Cleans up stale `/tmp` files
3. Removes duplicates
4. And more
## Motivation
 
In other words, you used to have to program your computer to do things like clean up after itself. Now it does it for you — but it always did; it just required a system programmer to go write a cron job. There is something nostalgic about building one for yourself in a language you like (Go).
 
Zig reminds me a lot of C — not in content, but in attitude. That's a topic for another essay.
 
