from tempfile import NamedTemporaryFile
from unittest import TestCase

from io import StringIO

from tf2_bind_gen import LogParser, UserConnected, UserDisconnected, KillMsg


class BingGenTestCase(TestCase):
    @staticmethod
    def _make_parser() -> LogParser:
        # binds_fp = open("../binds.txt", encoding='utf-8', errors='ignore')
        log_file = "test.log"

        with open(log_file, "w+") as log_fp:
            log_fp.write("""
KILLER connected
KILLER killed VICTIM with scattergun. (crit)
""")
        binds_fp = StringIO("""
        p:{player} v:{victim} t:{total} w:{weapon}
        "[scattergun] scattergun p:{player} v:{victim} t:{total} w:{weapon}
        """)
        cfg_fp = NamedTemporaryFile(mode="w+")
        stats_fp = NamedTemporaryFile(mode="w+")
        parser = LogParser(log_file, cfg_fp, binds_fp, stats_fp)
        return parser


class TestStringParser(BingGenTestCase):

    def test_strings(self):
        player = "KILLER"
        victim = "VICTIM"
        parser = self._make_parser()
        cases = (
            ('KILLER connected\n', 'connected'),
            ('KILLER killed VICTIM with pep_pistol. (crit)\n',
             KillMsg(player=player, victim=victim, weapon="pep_pistol", crit=". (crit)", total=1)),
            ('KILLER killed VICTIM with scattergun.\n',
             KillMsg(player=player, victim=victim, weapon="scattergun", crit=".", total=2)),
            ('KILLER killed VIC;TIM with scattergun.\n',
             KillMsg(player=player, victim="VICTIM", weapon="scattergun", crit=".", total=3)),
            ('Disconnecting from abandoned match server', 'disconnected')
        )

        for i, t in enumerate(cases):
            string, expected = t
            try:
                result = parser.parse_log(string)
            except UserDisconnected:
                parser.disconnected()
                self.assertEqual(parser.username, None)
            except UserConnected as u:
                parser.connected(u.username)
                self.assertEqual(parser.username, u.username)
            else:
                e = cases[i][1]
                self.assertEqual(result, e)
