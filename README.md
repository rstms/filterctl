# filterctl

The rspamd classifier on this mailserver adds an 'X-Spam-Score' header to each
message.  This header value ranges between -100.0 and +100.0, with higher
numbers indicating more spam characteristics.

This rspam-classes filter adds an 'X-Spam-Class' header value based on a list
of class names, each associated with a max score value.  Class names may then
be used for message filtering in the email client.

Each email user may customize the classes and thresholds used for their own
account using this email based command interface.  Commands are executed by
sending a message to 'filterctl@your_domain.com' with the command and
arguments as the 'Subject' line.  The message body is ignored.  A reply
message is sent for each command containing output and status.

Subject Line Commands:
------------------------------------------------------------------------------
list 

Return the complete set of rspamd class names and threshold values for the
sender address.

------------------------------------------------------------------------------
delete [CLASS ...]

Delete rspamd filter classes. If no CLASS names are specified, all classes
for the sender address are deleted.  Optionally, one or more CLASS names may
be provided to delete specific classes from the configuration.

------------------------------------------------------------------------------
reset [CLASS=THRESHOLD ...]

Replace the set of rspamd class thresholds with a new set provided as
arguments.  Each class name has a threshold value.  The threshold values set
the upper limit for each class.  Any number of classes may be defined.
If no class specifications are provided, default values will be used.

------------------------------------------------------------------------------
set CLASS=THRESHOLD

Add or update a single class name and threshold value.
CLASS is an identifier string.
THRESHOLD is a floating point number.

------------------------------------------------------------------------------
version 

Outputs program name, version, rspamd_classes library version, uid, and gid.

------------------------------------------------------------------------------
usage 

Output this message

------------------------------------------------------------------------------

