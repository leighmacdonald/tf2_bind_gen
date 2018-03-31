import logging
import re
import sqlite3
from collections import defaultdict
from datetime import datetime
from enum import Enum
from io import StringIO
from os import environ
from random import choice

# noinspection PyUnresolvedReferences
from os.path import isfile, join

logger = logging.getLogger("bind_gen")


class Class(Enum):
    SCOUT = "scout"
    SOLDIER = "soldier"
    PYRO = "pyro"
    DEMO = "demo"
    HEAVY = "heavy"
    ENGINEER = "engineer"
    MEDIC = "medic"
    SNIPER = "sniper"
    SPY = "spy"
    MULTI = "multi"


ID_MAP = dict()
PLAYERS = set()
PLAYER_CLASSES = {}
WEAPON_MAP = {
    Class.MULTI: {"fryingpan", "ham_shank", "necro_smasher", "nonnonviolent_protest", "pistol", "telefrag", "world",
                  "reserve_shooter"},
    Class.SCOUT: {"atomizer", "bat", "force_a_nature", "pep_pistol", "pistol_scout", "sandman", "scattergun",
                  "shortstop", "soda_popper", "wrap_assassin"},
    Class.SOLDIER: {
        "airstrike", "blackbox", "cow_mangler", "disciplinary_action", "liberty_launcher", "market_gardener",
        "quake_rl", "rocketlauncher_directhit", "shotgun_soldier", "tf_projectile_rocket", "unique_pickaxe_escape"
    },
    Class.PYRO: {"ai_flamethrower", "back_scratcher", "backburner", "deflect_promode", "deflect_rocket", "degreaser",
                 "detonator", "dragons_fury", "dragons_fury_bonus", "flamethrower", "flaregun", "hot_hand",
                 "phlogistinator", "powerjack", "rainblower", "scorch_shot", "shotgun_pyro", "sledgehammer",
                 "the_maul"},
    Class.DEMO: {
        "bottle", "demokatana", "iron_bomber", "loch_n_load", "loose_cannon", "loose_cannon_impact",
        "quickiebomb_launcher", "sticky_resistance", "tf_projectile_pipe", "tf_projectile_pipe_remote", "ullapool_caber"
    },
    Class.HEAVY: {
        "brass_beast", "family_business", "fists", "iron_curtain", "long_heatmaker", "minigun", "natascha",
        "steel_fists", "tomislav", "warrior_spirit"
    },
    Class.ENGINEER: {
        "frontier_justice", "obj_minisentry", "obj_sentrygun", "obj_sentrygun2", "obj_sentrygun3", "rescue_ranger",
        "robot_arm", "robot_arm_blender_kill", "robot_arm_combo_kill", "shotgun_primary", "wrangler_kill", "wrench",
        "wrench_jag"
    },
    Class.MEDIC: {
        "amputator", "blutsauger", "bonesaw", "crusaders_crossbow", "proto_syringe", "syringegun_medic", "taunt_medic",
        "ubersaw"
    },
    Class.SNIPER: {
        "awper_hand", "bazaar_bargain", "bushwacka", "machina", "player_penetration", "pro_rifle", "pro_smg", "smg",
        "shooting_star", "sniperrifle", "sydney_sleeper", "tf_projectile_arrow", "the_classic", "tribalkukri"
    },
    Class.SPY: {
        "ambassador", "big_earner", "black_rose", "diamondback", "enforcer", "eternal_reward", "knife", "kunai",
        "letranger", "revolver", "sharp_dresser", "spy_cicle"
    }
}


def get_id(player):
    try:
        return ID_MAP[player]
    except KeyError:
        PLAYERS.add(player)
        return player


def get_class(weapon) -> Class:
    for class_name, weapons in WEAPON_MAP.items():
        if weapon in weapons:
            return class_name
    return Class.MULTI


def init_db(conn: sqlite3.Connection, drop=False):
    tables = ("kills",)
    cur = conn.cursor()
    if drop:
        for table_name in tables:
            cur.execute("SELECT name FROM sqlite_master WHERE type='table' AND name=?", (table_name,))
            res = cur.fetchone()
            if res:
                cur.execute("DROP TABLE {};".format(table_name))
        conn.commit()
    cur.execute("SELECT name FROM sqlite_master WHERE type='table' AND name IN (?)", tables)
    res = cur.fetchall()
    if len(res) != len(tables):
        cur.execute("""
            CREATE TABLE kills
            (
                kill_id INTEGER PRIMARY KEY AUTOINCREMENT,
                steam_id TEXT NOT NULL,
                weapon TEXT NOT NULL,
                is_crit BOOLEAN,
                created_on TIMESTAMP
            );
        """)
        conn.commit()
    for table_name in tables:
        cur.execute("SELECT name FROM sqlite_master WHERE type='table' AND name=?", (table_name,))
        res = cur.fetchone()
        if not res:
            raise ValueError("Failed to find table: {}".format(table_name))
    cur.close()


class BindGenExc(Exception):
    pass


class UserConnected(BindGenExc):
    def __init__(self, username):
        self.username = username


class IDMapping(BindGenExc):
    def __init__(self, username, steam_id):
        self.username = username
        self.steam_id = steam_id


class UserDisconnected(UserConnected):
    pass


class KillMsg(object):
    def __init__(self, player, victim, weapon, crit, total=0):
        self.player = player
        # TF2 has no escaping in aliases...
        # Don't allow users to inject via their name
        for c in (";", "\"", "\'"):
            victim = victim.replace(c, "")
        self.victim = victim
        self.victim_steam_id = get_id(victim)
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

    def __eq__(self, other):
        return self.__dict__ == other.__dict__


class StatLogger(object):
    def __init__(self, conn: sqlite3.Connection):
        self.conn = conn
        self.stats = defaultdict(int)
        self.cursor = conn.cursor()

    def write(self):
        self.conn.commit()

    def get(self, steam_id):
        self.cursor.execute("""SELECT count(steam_id) FROM kills WHERE steam_id = ?""", (steam_id,))
        res = self.cursor.fetchone()
        return res[0] if res else 0

    def increment(self, kill_msg: KillMsg):
        try:
            self.cursor.execute("""INSERT INTO kills (steam_id, weapon, is_crit, created_on) VALUES (?, ?, ?, ?)""",
                                (kill_msg.victim_steam_id, kill_msg.weapon, kill_msg.crit, datetime.now()))
        except KeyError:
            logger.warning("No steam_id associated with {}. Please run the 'status' console command".format(
                kill_msg.player))
        else:
            self.write()
            self.cursor.execute("SELECT count(steam_id) FROM kills WHERE steam_id = ?", (kill_msg.victim_steam_id,))
            res = self.cursor.fetchone()
            return res[0]

    def migrate_player(self, player_name):
        try:
            new_id = ID_MAP[player_name]
            self.cursor.execute("UPDATE kills SET steam_id = ? WHERE steam_id = ?", (new_id, player_name,))
            self.conn.commit()
            count = self.cursor.rowcount
            if player_name in PLAYERS:
                PLAYERS.remove(player_name)
            logger.debug("Migrated {} kill entries to steam_id".format(count))
        except KeyError:
            pass


class LogParser(object):
    #  NAME killed NAME with GUN.
    _re_kill = re.compile(r"^(.+?)\skilled\s(.+?)\swith\s(.+)(\.|\. \(crit\))$")

    #  NAME connected
    _re_connected = re.compile(r"^(\S+)\sconnected$")

    _status_id_re = re.compile(r'"(.+?)"\s+(\[U:\d+:\d+\]|STEAM_\d:\d:\d+)')  # STEAM_1:0:159598523

    #  Disconnecting from abandoned match server
    _re_disconnect = re.compile(r"(^Disconnecting from abandoned match server$|\(Server shutting down\)$)")

    _re_bind_key = re.compile(r"^\[(.+?)\](.+?)$")

    def __init__(self, stat_logger: StatLogger, log_path, config_path: str, binds_fp):
        """

        b = open("../binds.txt", encoding='utf-8', errors='ignore')
        LogParser("", "", b, "")

        :param log_path:
        :param binds_fp:
        """
        self.log_path = log_path
        self.config_path = config_path
        self.username = None
        self.default_bind_key = "generic"
        self.templates = self.read_binds(binds_fp)
        self.stats = stat_logger

    def parse_log(self, line):
        if self.username is None:
            m = self._re_connected.search(line)
            if m:
                username = m.groups()[0]
                raise UserConnected(username)
        if self._re_disconnect.match(line):
            raise UserDisconnected(self.username)
        match = self._re_kill.search(line)
        if match:
            msg = KillMsg(*match.groups())
            killed_class = get_class(msg.weapon)
            if killed_class != Class.MULTI:
                PLAYER_CLASSES[msg.player] = killed_class
                logger.debug("Assigning {} to class {}".format(msg.player, killed_class))
            if msg.player == self.username:
                msg.total = self.stats.increment(msg)
                logger.debug(msg)
                return msg
        status = self._status_id_re.search(line)
        if status:
            values = status.groups()
            raise IDMapping(*values)

    def read_binds(self, fp: StringIO):
        fp.seek(0)
        found = 0
        binds = defaultdict(list)
        for line in fp.readlines():
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
        msg_str = self.gen_message(msg)
        logger.info(msg_str)
        with open(self.config_path, mode="w+", encoding='utf-8', errors='ignore') as log_cfg:
            log_cfg.write('echo "Loaded log_parser.cfg"\n')
            alias = '''alias bind_gen "say {} "'''.format(msg_str)
            logger.debug(alias)
            log_cfg.write(alias + "\n")

    def gen_message(self, msg: KillMsg):
        if msg.victim in PLAYER_CLASSES and self.templates[PLAYER_CLASSES[msg.victim].value]:
            template = choice(self.templates[PLAYER_CLASSES[msg.victim].value])
        elif msg.key in self.templates:
            template = choice(self.templates[msg.key])
        else:
            template = choice(self.templates[self.default_bind_key])
        output_str = template.format(victim=msg.victim, player=msg.player, weapon=msg.weapon,
                                     total=msg.total)
        return output_str

    def disconnected(self):
        logger.info("Disconnected from server")
        self.username = None
        self.stats.write()

    def connected(self, username):
        logger.info("Connected with username: {}".format(username))
        self.username = username

    def start(self):
        for line in self.tail():
            if line:
                self.handle_line(line)

    def stop(self):
        logger.info("Shutting down...")
        if self.username:
            self.disconnected()

    def read_file(self, log_file):
        for line in open(log_file, encoding='utf-8', errors='ignore').readlines():
            self.handle_line(line)

    def handle_line(self, line):
        try:
            msg = self.parse_log(line)
        except UserDisconnected:
            self.disconnected()
        except UserConnected as u:
            self.connected(u.username)
        except IDMapping as mapping:
            ID_MAP[mapping.username] = mapping.steam_id
            self.stats.migrate_player(mapping.username)
            logger.debug("Set SteamID {} -> {}".format(mapping.username, mapping.steam_id))
        else:
            if msg:
                self.write_cfg(msg)

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


def run(args):
    logging.basicConfig(level=logging.DEBUG if args.debug else logging.INFO,
                        format="[TF2BindGen] [%(levelname)s] %(message)s")
    connection = sqlite3.connect(args.db)
    try:
        init_db(connection, drop=False)
        b_fp = open(args.binds, encoding='utf-8', errors='ignore')
    except IOError as e:
        logger.exception(e)
    else:
        stats = StatLogger(connection)
        parser = LogParser(stats, args.log_path, args.config_path, b_fp)
        if args.test:
            parser.read_file(args.log_path)
        else:
            try:
                parser.start()
            except KeyboardInterrupt:
                parser.stop()
    finally:
        connection.close()


if __name__ == "__main__":
    import argparse

    try:
        program_files_path = environ['PROGRAMFILES(X86)']
    except KeyError:
        program_files_path = environ['PROGRAMFILES']
    log_path_default = join(program_files_path, r"Steam\steamapps\common\Team Fortress 2\tf\console.log")
    config_path_default = join(program_files_path, r"Steam\steamapps\common\Team Fortress 2\tf\cfg\log_parser.cfg")
    arg_parser = argparse.ArgumentParser(description='TF2 Log Tail Parser, Stat Tracker and Text Bind Generator')
    arg_parser.add_argument('--log_path', default=log_path_default,
                            help="Path to console.log generated by TF2 (default: {})".format(log_path_default))
    arg_parser.add_argument('--config_path', default=config_path_default,
                            help="Path to the .cfg file to be generated (default: {}".format(config_path_default))
    arg_parser.add_argument('--db', default="stats.db",
                            help="Path to the database file (default: {})".format("stats.db"))
    arg_parser.add_argument('--test', action='store_true', help="Test parsing your existing log files (default: False)")
    arg_parser.add_argument('--binds', default="binds.txt", help="Path to your custom binds file. (default: binds.txt)")
    arg_parser.add_argument('--debug', action='store_true', help="Set the logging level to debug. (default: False)")
    run(arg_parser.parse_args())
