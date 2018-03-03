import re
from collections import namedtuple
from os import environ
from os.path import isfile, join
import logging

KillMsg = namedtuple("KillMsg", ['player', 'victim', 'weapon'])


class LogParser(object):
    #  NAME killed NAME with GUN.
    _re_kill = re.compile(r"^(.+)\skilled\s(.+)\swith\s(.+)\.|\. \(crit\)$")

    #  NAME connected
    _re_connected = re.compile(r"^(.+)\sconnected$")

    #  Disconnecting from abandoned match server
    _re_disconnect = re.compile(r"^Disconnecting from abandoned match server$")

    def __init__(self, log_path, cfg_path, bind_key):
        self.log_path = log_path
        self.cfg_path = cfg_path
        self.bind_key = bind_key
        self.username = None

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
        else:
            match = self._re_kill.search(line)
            if match:
                msg = KillMsg(*match.groups())
                if msg.player == self.username:
                    logging.info(msg)
                    self.write_cfg(msg)

    def write_cfg(self, msg: KillMsg):
        with open(self.cfg_path, 'w+') as cfg:
            cfg.write('echo "Loaded log_parser.cfg"\n')
            cfg.write('bind {} "say Get rekt {}!"\n'.format(self.bind_key, msg.victim))

    def start(self):
        for line in self.tail():
            self.parse_log(line)

    def read_file(self, log_file):
        for line in open(log_file).readlines():
            self.parse_log(line)

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
    program_files_path = environ['PROGRAMFILES(X86)']
    log_path_default = join(program_files_path, r"Steam\steamapps\common\Team Fortress 2\tf\console.log")
    config_path_default = join(program_files_path, r"Steam\steamapps\common\Team Fortress 2\tf\cfg\log_parser.cfg")
    parser = argparse.ArgumentParser(description='TF2 Log Tail Parser')
    parser.add_argument('--log_path', default=log_path_default)
    parser.add_argument('--config_path', default=config_path_default, help="Path to the .cfg file to be generated")
    parser.add_argument('--bind_key', default="f1", help="Keyboard shortcut used for chat bind")
    parser.add_argument('--test', type=bool, default=False, help="Test parsing your existing log files")
    args = parser.parse_args()
    parser = LogParser(args.log_path, args.config_path, args.bind_key)
    if args.test:
        parser.read_file(log_path_default)
    else:
        parser.start()
