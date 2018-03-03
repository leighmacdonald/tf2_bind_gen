import json
import re
from collections import namedtuple, defaultdict
from os import environ
from os.path import isfile, join, exists
import logging
from random import choice

KillMsg = namedtuple("KillMsg", ['player', 'victim', 'weapon', 'crit', 'total'])


class StatLogger(object):
    def __init__(self, stats_file, write_every=5):
        self.stats_file = stats_file
        self.writer_count = 0
        self.write_every = write_every
        self.stats = defaultdict(int)

    def write(self):
        with open(self.stats_file, "w", encoding='utf8') as log:
            json.dump(self.stats, log)

    def read(self):
        if not exists(self.stats_file):
            return False
        with open(self.stats_file, encoding='utf8') as log:
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

    def __init__(self, log_path, cfg_path, bind_key, binds_file, stats_file):
        self.log_path = log_path
        self.cfg_path = cfg_path
        self.bind_key = bind_key
        self.username = None
        self.default_bind_key = "generic"
        self.templates = self.read_binds(binds_file)
        self.stats = StatLogger(stats_file)

    def parse_log(self, line):
        if self.username is None:
            m = self._re_connected.search(line)
            if m:
                self.username = m.groups()[0]
                logging.info("Connected with username: {}".format(self.username))
                return
        elif self._re_disconnect.match(line):
            self.username = None
            logging.info("Disconnected from server")
            self.stats.write()
        else:
            match = self._re_kill.search(line)
            if match:
                msg_args = list(match.groups())
                msg_args[-1] = True if "crit" in msg_args[-1] else False
                msg_args.append(self.stats.increment(msg_args[1]))
                msg = KillMsg(*msg_args)
                if msg.player == self.username:
                    logging.info(msg)
                    self.write_cfg(msg)
                    return msg

    def read_binds(self, file_name):
        found = 0
        binds = defaultdict(list)
        for line in open(file_name).readlines():
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
        logging.info("Loaded {} binds".format(found))
        return binds

    def write_cfg(self, msg: KillMsg):
        with open(self.cfg_path, 'w+') as cfg:
            cfg.write('echo "Loaded log_parser.cfg"\n')
            alias = '''alias bind_gen "say {} "\n'''.format(self.gen_message(msg))
            logging.debug(alias)
            cfg.write(alias)

    def gen_message(self, msg: KillMsg):
        key = "{}.crit".format(msg.weapon) if msg.crit else "{}".format(msg.weapon)
        try:
            template = choice(self.templates[key])
        except IndexError:
            template = choice(self.templates[self.default_bind_key])
        output_str = template.format(victim=msg.victim, player=msg.player, weapon=msg.weapon,
                                     total=msg.total)
        return output_str

    def start(self):
        for line in self.tail():
            self.parse_log(line)

    def stop(self):
        logging.info("Shutting down...")
        self.stats.write()

    def read_file(self, log_file):
        for line in open(log_file, encoding='utf8').readlines():
            msg = self.parse_log(line)
            if msg:
                print(self.gen_message(msg))

    def tail(self):
        first_call = True
        while True:
            try:
                with open(self.log_path) as log_file:
                    if first_call:
                        log_file.seek(0, 2)
                        first_call = False
                    latest_data = log_file.read()
                    while True:
                        if '\n' not in latest_data:
                            latest_data += log_file.read()
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

    logging.basicConfig(level=logging.DEBUG)
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
    parser.add_argument('--bind_key', default="f2", help="Keyboard shortcut used for chat bind (default: f2)")
    parser.add_argument('--test', action='store_true', help="Test parsing your existing log files (default: False)")
    parser.add_argument('--binds', default="binds.txt", help="Path to your custom binds file. (default: binds.txt)")
    parser.add_argument('--stats', default="stats.json", help="Path to your stats file. (default: stats.json)")
    args = parser.parse_args()

    parser = LogParser(args.log_path, args.config_path, args.bind_key, args.binds, args.stats)
    if args.test:
        parser.read_file(log_path_default)
    else:
        try:
            parser.start()
        except KeyboardInterrupt:
            parser.stop()
