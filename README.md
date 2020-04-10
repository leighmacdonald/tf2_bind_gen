TF2 Dynamic Chat Bind Generator
===============================

NOTE: For the previous python based version, see the [python](https://github.com/leighmacdonald/tf2_bind_gen/tree/python) branch.

What Is This?
-------------

This script can be used to generate new chat binds on the fly based on certain events that happen such as
killing a player. This data is being parsed from the TF2 console log output. This means that it will update your
binds while in game automatically without having to leave. It will randomly choose a chat template from a
list of your own template definitions. These definitions can also be customized and grouped based on things
like if a kill is from a certain weapon, or a crit kill.


Installation
------------

The recommended way of running the tool is using the pre built executables available on the releases page.

Versions:

- Latest 2.x Soon. Build from source if you can't wait.
- Legacy 1.x [Download](https://github.com/leighmacdonald/tf2_bind_gen/releases/download/v1.6/tf2_bind_gen-v1.6.zip).

If you don't trust me, or you want to edit or view the code yourself, you must have the below requirements met:

The application only requires [golang](https://golang.org/dl/) to be installed to build. Any semi-recent version should
be acceptable.

Download the source [zip file](https://github.com/leighmacdonald/tf2_bind_gen/archive/master.zip) and extract it anywhere you want. It 
does not need to be, and should not be, in the tf2 directory to work.

You can of course clone the repo as well if you have a git client installed.

Usage
-----

For this to function you must have a key bound to reload the newly updated
binds generated on the fly. Add a bind like the following to your autoexec.cfg file or the appropriate cfg 
file for your configuration, replacing the bound key to one of your choosing.

    bind f1 "exec log_parser; bind_gen"
    
If you don't add a key to reload the log_parser.cfg file you will never get updated player names. So this is 
very important and probably the most likely area to encounter a problem with the script.

Additionally, you must also add the following launch options to TF2:

     -condebug -conclearlog
     
This will enable the required output of tf2 log messages to the default log path. -conclearlog is optional but
it will clear the log file upon startup so the log file should not get too large.

    Usage:
      bind_generator [flags]
    
    Flags:
          --config string   config file (default is ./config.yaml)
      -d, --debug           Enable debug output
      -h, --help            help for bind_generator
    
Once running you can use your two binds to reload the bind_generator.cfg and to execute your bind.

Customizing Your Sick Memes
---------------------------

You can customize the binds used by creating or editing the binds.txt file. The format is 1 bind per line and 
has the following variables which can be used for substitutions: 
 
- {{ .Player }} - Your player name  (previously {player})
- {{ .Victim }} - The name of the person you killed (previously {victim})
- {{ .Weapon }} - The weapon you killed them with. (only console names so: "tf_projectile_rocket" and not "Rocket Launcher") (previously {weapon})
- {{ .Kills }} - The number of times you've killed a player with that name. Doesnt currently track Steam ID. (previously {kills})

Variables that start with a `$` are "dynamic" and can do things like make HTTP requests to external sites. Keep in mind that these 
are subject to high latency since the 3rd party services can take varying amounts of time to respond.

- $google_result - Will query google and pick a result to use for a link at random. If < 10 results are returned, the 1st result will always be used.
If there is > 10 results, a random one will be picked from the first 10 results.  

The standard template variables `{{ ... }}` are always executed / substituted before the dynamic ` $...` variables. 

NOTE: If migrating from 1.x, you can easily just search and replace all your variables to the new formats.

Some examples are below:

    [generic] Get rekt {victim} That makes it {total}! :)
    [generic] Why so mad {victim}?
    [generic] {player} rekt {victim} LOL!
    [market_gardener.crit] Try looking up next time {victim}!
    [tf_projectile_rocket.crit] EZ Crit! Thanks {victim}!
    [tf_projectile_rocket] Thanks for the farm {victim}!
    [world] {player} > world > {victim}
    [scout] lol {victim} is bad at scout!
    [sniper] lol {victim} is bad at sniper!
    [engineer] lol {victim} is bad at engy!
    [generic] See $google_result for more hilarious deaths by {{ .Victim }}!

You can specify binds for specific weapon kills and whether they are a crit or not. These keys correspond to the 
weapon names you see in the console kill log. I don't have a list of all of these, you can check your console logs if 
you don't know the name of something you want to use. See above for examples of crit and non-crit weapon binds. If no
[key] is defined, it will be defined as [generic] for you by default.

Additionally, you can create binds for specific classes. Be aware **PlayerClass specific binds take precedence** over all
other bind types. On-top of this, they can only be so specific because the TF2 console log does not specify which class
a user has switched to. This means that we can only track the class the user what when they killed another player with
a non multi class weapon. In other words, there is no way to tell what class a player is until they have killed someone
with a unique weapon type. If they change classes after killing somebody they will still be known to be the previous class
until they again kill another player with a unique weapon type.

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
