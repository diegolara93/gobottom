# BottomGo
A graphical terminal interface for system monitoring made in Go inspired by bottom(https://github.com/ClementTsang/bottom), which is written in Rust.
![Alt text](/screenshots/Screenshot%202024-12-29%20012210.png?raw=true)
Any suggestions on what I should add or change are welcome! I just started working on this so I know it is still lacking a lot.
## How It's Made:

**Made with:** Go, BubbleTea, NtCharts, GoPsutil, LipGloss, BubbleZone, Bubbles
I originally made this to try and learn more about TUI's, originally prototyping this same project in Rust with Ratatui(https://ratatui.rs/),
I found BubbleTea to be feel more intuitive though and felt myself getting more done using it. 
A lot of refactoring needs to be done as I just threw stuff together as I went.



## Known Issues:
Due to issues with GoPsutil, specifically https://github.com/giampaolo/psutil/issues/906 on FreeBSD based operating systems, such as MacOS darwin,
disks will display all partitions of every disk rather than just the physical disks of your machine.



