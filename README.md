TF2 Bind Generator
==================

Installation
------------

The script requires [python](https://www.python.org/downloads/) to be installed. I have only tested
with 3.3+, but 2.7 may also work.


Usage
-----

For this to function you must have a key bound to reload the newly updated
binds generated on the fly. Add a bind like the following to your autoexec.cfg file, replacing
the bound key to one of your choosing.

    bind f1 "exec log_parser"
    
If you don't add a key to reload the log_parser.cfg file you will never get updated player names.

    usage: tf2_bind_gen.py [-h] [--log_path LOG_PATH] [--config_path CONFIG_PATH]
                       [--bind_key BIND_KEY] [--test TEST]

    TF2 Log Tail Parser
    
    optional arguments:
      -h, --help            show this help message and exit
      --log_path LOG_PATH
      --config_path CONFIG_PATH
                            Path to the .cfg file to be generated
      --bind_key BIND_KEY   Keyboard shortcut used for chat bind
      --test TEST           Test parsing your existing log files
      
The default settings should work for most users. If you want to change the key that the bind gets bound
too set the --bind_key option to the key of your choosing. For example:

    tf2_bind_gen.py --bind_key f4
    
Once running you can use your two binds to reload the log_parser.cfg and to execute your bind.
