# README

1. Set `GIHTUB_ACCESS_TOKEN=<your access token>` in bash shell
   You can create a new access token with permission `user` at https://github.com/settings/tokens.
2. Run `gh-followers` to show your current followers list:
    - this will recorded current followers into `~/.config/gh-followers/followers.*`.
    - if previous files existed, further diff will be executed, list unfollowed/newfollowed users since then.
    
