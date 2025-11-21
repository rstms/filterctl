# filterctl
filterctl is a mail-based command processor for user management of spam class filter settings.
It interprets the subject line as a command, executes it, and sends the output as the body
of a new email message sent back to the sender address.
It relies on the mailserver configuration to guarantee that only local mail originating from
a local account via an authorized secure SMTPS session is accepted for the the filterctl user.
In this way it relies on the security of the mailserver's auth mechanism to control access to
the commands.  The sender address is verified as an authorized local user.
Command actions issue API requests to filterctld running at http://localhost:2016/filterctl
