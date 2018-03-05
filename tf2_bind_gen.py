import json
import re
from collections import defaultdict
from os import environ
from os.path import isfile, join, exists
import logging
from random import choice

logger = logging.getLogger("bind_gen")


class KillMsg(object):
    def __init__(self, player, victim, weapon, crit, total=0):
        self.player = player
        self.victim = victim
        self.weapon = weapon
        self.crit = True if "crit" in crit else False
        self.total = total

    @property
    def key(self):
        if self.crit:
            return "{}.crit".format(self.weapon)
        else:
            return self.weapon

    def __str__(self):
        return "victim: {} weapon: {} crit: {}".format(self.victim, self.weapon, self.crit)


class StatLogger(object):
    def __init__(self, stats_file, write_every=5):
        self.stats_file = stats_file
        self.writer_count = 0
        self.write_every = write_every
        self.stats = defaultdict(int)

    def write(self):
        with open(self.stats_file, "w", encoding='utf-8', errors='ignore') as log:
            json.dump(self.stats, log)

    def read(self):
        if not exists(self.stats_file):
            return False
        with open(self.stats_file, encoding='utf-8', errors='ignore') as log:
            x = json.load(log)
            self.stats.update(x)

    def get(self, user_name):
        return self.stats[user_name]

    def increment(self, user_name):
        self.stats[user_name] += 1
        self.writer_count += 1
        if self.writer_count >= self.write_every:
            self.write()
            self.writer_count = 0
        return self.stats[user_name]


class LogParser(object):
    #  NAME killed NAME with GUN.
    _re_kill = re.compile(r"^(.+)\skilled\s(.+)\swith\s(.+)(\.|\. \(crit\))$")

    #  NAME connected
    _re_connected = re.compile(r"^(.+)\sconnected$")

    #  Disconnecting from abandoned match server
    _re_disconnect = re.compile(r"(^Disconnecting from abandoned match server$|\(Server shutting down\)$)")

    _re_bind_key = re.compile(r"^\[(.+?)\](.+?)$")

    def __init__(self, log_path, cfg_path, binds_file, stats_file):
        self.log_path = log_path
        self.cfg_path = cfg_path
        self.username = None
        self.default_bind_key = "generic"
        self.templates = self.read_binds(binds_file)
        self.stats = StatLogger(stats_file)

    def parse_log(self, line):
        if self.username is None:
            m = self._re_connected.search(line)
            if m:
                self.username = m.groups()[0]
                logger.info("Connected with username: {}".format(self.username))
                return
        elif self._re_disconnect.match(line):
            self.username = None
            logger.info("Disconnected from server")
            self.stats.write()
        else:
            match = self._re_kill.search(line)
            if match:
                msg = KillMsg(*match.groups())
                if msg.player == self.username:
                    msg.total = self.stats.increment(msg.victim)
                    logger.debug(msg)
                    self.write_cfg(msg)
                    return msg

    def read_binds(self, file_name):
        found = 0
        binds = defaultdict(list)
        for line in open(file_name, encoding='utf-8', errors='ignore').readlines():
            real_line = line.strip()
            if real_line not in binds:
                match_key = self._re_bind_key.search(real_line)
                if match_key:
                    key, raw_msg = match_key.groups()
                    msg = raw_msg.strip()
                else:
                    key = self.default_bind_key
                    msg = real_line
                binds[key].append(msg)
                found += 1
        logger.info("Loaded {} binds".format(found))
        return binds

    def write_cfg(self, msg: KillMsg):
        with open(self.cfg_path, 'w+', encoding='utf-8', errors='ignore') as cfg:
            cfg.write('echo "Loaded log_parser.cfg"\n')
            alias = '''alias bind_gen "say {} "'''.format(self.gen_message(msg))
            logger.debug(alias)
            cfg.write(alias + "\n")

    def gen_message(self, msg: KillMsg):
        try:
            template = choice(self.templates[msg.key])
        except IndexError:
            template = choice(self.templates[self.default_bind_key])
        output_str = template.format(victim=msg.victim, player=msg.player, weapon=msg.weapon,
                                     total=msg.total)
        return output_str

    def start(self):
        for line in self.tail():
            self.parse_log(line)

    def stop(self):
        logger.info("Shutting down...")
        self.stats.write()

    def read_file(self, log_file):
        for line in open(log_file, encoding='utf-8', errors='ignore').readlines():
            msg = self.parse_log(line)
            if msg:
                logger.info(self.gen_message(msg))

    def tail(self):
        first_call = True
        while True:
            try:
                with open(self.log_path, encoding='utf-8', errors='ignore') as log_file:
                    if first_call:
                        log_file.seek(0, 2)
                        first_call = False
                    latest_data = log_file.read()
                    while True:
                        if '\n' not in latest_data:
                            try:
                                latest_data += log_file.read()
                            except UnicodeDecodeError as err:
                                logger.exception(err)
                                continue
                            if '\n' not in latest_data:
                                yield ''
                                if not isfile(self.log_path):
                                    break
                                continue
                        latest_lines = latest_data.split('\n')
                        if latest_data[-1] != '\n':
                            latest_data = latest_lines[-1]
                        else:
                            latest_data = log_file.read()
                        for line in latest_lines[:-1]:
                            yield line + '\n'
            except IOError:
                yield ''


if __name__ == "__main__":
    import argparse

    try:
        program_files_path = environ['PROGRAMFILES(X86)']
    except KeyError:
        program_files_path = environ['PROGRAMFILES']
    log_path_default = join(program_files_path, r"Steam\steamapps\common\Team Fortress 2\tf\console.log")
    config_path_default = join(program_files_path, r"Steam\steamapps\common\Team Fortress 2\tf\cfg\log_parser.cfg")
    parser = argparse.ArgumentParser(description='TF2 Log Tail Parser')
    parser.add_argument('--log_path', default=log_path_default,
                        help="Path to console.log generated by TF2 (default: {})".format(log_path_default))
    parser.add_argument('--config_path', default=config_path_default,
                        help="Path to the .cfg file to be generated (default: {}".format(config_path_default))
    parser.add_argument('--test', action='store_true', help="Test parsing your existing log files (default: False)")
    parser.add_argument('--binds', default="binds.txt", help="Path to your custom binds file. (default: binds.txt)")
    parser.add_argument('--stats', default="stats.json", help="Path to your stats file. (default: stats.json)")
    parser.add_argument('--debug', action='store_true', help="Set the logging level to debug. (default: False)")
    args = parser.parse_args()

    logging.basicConfig(level=logging.DEBUG if args.debug else logging.INFO,
                        format="[TF2BindGen] [%(levelname)s] %(message)s")

    parser = LogParser(args.log_path, args.config_path, args.binds, args.stats)
    if args.test:
        parser.read_file(log_path_default)
    else:
        try:
            parser.start()
        except KeyboardInterrupt:
            parser.stop()
