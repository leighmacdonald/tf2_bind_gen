TF2 Bind Generator
==================

Installation
------------

The script only requires [python](https://www.python.org/downloads/) to be installed. It uses no external dependencies.
I have only tested with python 3.3+, but 2.7 may also work.


Usage
-----

For this to function you must have a key bound to reload the newly updated
binds generated on the fly. Add a bind like the following to your autoexec.cfg file or the appropriate cfg 
file for your configuration, replacing the bound key to one of your choosing. I would recommend not using autoexec.cfg
if possible due to not wanting to re-execute all of it each time, but it will work well enough for most users. Having a 
custom "binds.cfg" for example could be a good choice.

    bind f1 "exec log_parser"
    
If you don't add a key to reload the log_parser.cfg file you will never get updated player names. So this is 
very important and probably the most likely area to encounter a problem with the script.

Additionally, you must also add the following launch options to TF2:

     -condebug -conclearlog
     
This will enable the required output of tf2 log messages to the default log path. -conclearlog is optional but
it will clear the log file upon startup so the log file should not get too large.
 

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

Customizing Your Sick Memes
---------------------------

You can customize the binds used by creating or editing the binds.txt file. The format is 1 bind per line and 
has the following variables which can be used: player, victim, weapon. Some examples are below:

    Get rekt {victim}!
    [generic] Why so mad {victim}?
    [generic] {player} rekt {victim} LOL!
    [market_gardener.crit] Try looking up next time {victim}!
    [tf_projectile_rocket.crit] EZ Crit! Thanks {victim}!
    [tf_projectile_rocket] Thanks for the farm {victim}!
    [world] world > {victim}
    
You can specify binds for specific weapon kills and whether they are a crit or not. These keys correspond to the 
weapon names you see in the console kill log. I don't have a list of all of these, you can check your console logs if 
you don't know the name of something you want to use. See above for examples of crit and non-crit weapon binds. If no
[key] is defined, it will be defined as [generic] for you by default.

If you are using an [Australium weapon](https://wiki.teamfortress.com/wiki/Australium), there is no way to know if a 
kill is a crit or not, all kills are considered a crit according to the console log. So if you want to match a australium
weapon, make sure to always add ".crit" to the bind key configuration.

If you are using a file other than binds.txt you can change the default path to your binds by specifying the --binds 
flag value.
 
    tf2_bind_gen.py --binds binds_custom.txt

Will this get me VAC banned?
----------------------------

Short answer: No. 

Less short answer: No, all this script does is read and parse the tf2 log file which is a plain text file. 
It then write a new .cfg. There is no functionality to read live TF2 data from memory, attaching to 
running process or DLL injections happening at all. 

Example
-------

![Example 1](https://raw.githubusercontent.com/leighmacdonald/tf2_bind_gen/master/example/screen_1.jpg)
