# twitch-classifier

A simple framework/skeleton for running a naive bayes classifier on twitch messages

## About

Eventually this is planned to be more of an all in one project, but my dataset is currently pretty early and experimental, so currently this is more of a skeleton to plug your own data in. This code is also very early, with plans to be cleaned up massively and turned into a more userfriendly CLI

Currently the format for data is a simple csv containing a column of lines of chat, and a column for a numerical value, `0` being a negative comment, `1` for neutral, and `2` for positive, this may change in the future. By default this is loaded from `data.csv` in the root directory of the project

Example `data.csv`:

```csv
really love the stream! :),2
this stream sucks,0
```