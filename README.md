# BottomGo
A graphical terminal interface for system monitoring inspired by bottom(https://github.com/ClementTsang/bottom), which is written in Rust.

## How It's Made:

**Made with:** Go, BubbleTea, NtCharts, GoPsutil
I originally made this to try and learn more about TUI's, originally prototyping this same project in Rust with Ratatui(https://ratatui.rs/),
I found BubbleTea to be feel more intuitive though and felt myself getting more done using it. 
A lot of refactoring needs to be done as I just threw stuff together as I went.
A lot of the BubbleTea helper libraries are used for styling and other utilities such as mouse tracking. 



## Known Issues:
Due to issues with GoPsutil, specifically https://github.com/giampaolo/psutil/issues/906 on FreeBSD based operating systems, such as MacOS darwin,
disks will display all partitions of every disk rather than just the physical disks of your machine.



