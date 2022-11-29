Replica failure happens in logs found in manual

We however recommend viewing logs through use of linearity.xlsx - ESPECIALLY for autoclient, since there are a couple of thousands lines. One thing to make note of, is that sometimes in linearity.xlsx you may find a weird case of response returning from replica, before the replica has even logged that a client is requesting - this is not a bug, but merely a result of the timestamp being the exact same, and Excel not differentiating between them.

We therefore incur the reader to check if the timestamps are identical if a weird order appears - if they are, the order should be switched in your mind.