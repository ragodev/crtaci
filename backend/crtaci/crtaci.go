// Author: Milan Nikolic <gen2brain@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gen2brain/vidextr"
	"github.com/google/google-api-go-client/googleapi/transport"
	youtube "github.com/google/google-api-go-client/youtube/v3"
)

var (
	appName    = "crtaci-http"
	appVersion = "1.5"
)

type cartoon struct {
	Id             string `json:"id"`
	Character      string `json:"character"`
	Title          string `json:"title"`
	FormattedTitle string `json:"formattedTitle"`
	Episode        int    `json:"episode"`
	Season         int    `json:"season"`
	Service        string `json:"service"`
	Url            string `json:"url"`
	ThumbSmall     string `json:"thumbSmall"`
	ThumbMedium    string `json:"thumbMedium"`
	ThumbLarge     string `json:"thumbLarge"`
}

type character struct {
	Name     string `json:"name"`
	AltName  string `json:"altname"`
	AltName2 string `json:"altname2"`
	Duration string `json:"duration"`
	Query    string `json:"query"`
}

var characters = []character{
	{"atomski mrav", "", "", "medium", ""},
	{"a je to", "", "", "medium", "a je to crtani"},
	{"anđeoski prijatelji", "andjeoski prijatelji", "", "medium", ""},
	{"bananamen", "", "", "medium", ""},
	{"blinki bil", "", "блинки бил", "long", ""},
	{"blufonci", "", "", "medium", ""},
	{"bombončići", "bomboncici", "", "medium", ""},
	{"braća grim", "braca grim", "najlepse bajke", "long", ""},
	{"brzi gonzales", "", "", "medium", ""},
	{"čarli braun", "carli braun", "", "medium", ""},
	{"čarobni školski autobus", "carobni skolski autobus", "", "long", ""},
	{"čili vili", "cili vili", "", "medium", ""},
	{"cipelići", "cipelici", "", "medium", ""},
	{"denis napast", "", "", "long", ""},
	{"doživljaji šašave družine", "dozivljaji sasave druzine", "", "long", ""},
	{"droidi", "", "", "long", ""},
	{"duško dugouško", "dusko dugousko", "dusko 20dugousko", "medium", ""},
	{"džoni test", "dzoni test", "", "long", ""},
	{"elmer", "", "", "medium", "elmer crtani"},
	{"eustahije brzić", "eustahije brzic", "", "medium", ""},
	{"evoksi", "", "", "long", ""},
	{"generalova radnja", "", "", "medium", ""},
	{"grčka mitologija", "grcka mitologija", "", "long", "grcka mitologija crtani"},
	{"gustav", "gustavus", "", "medium", "gustavus crtani"},
	{"helo kiti", "", "", "medium", ""},
	{"hi men i gospodari svemira", "himen i gospodari svemira", "", "long", ""},
	{"inspektor radiša", "inspektor radisa", "", "medium", ""},
	{"iznogud", "", "", "medium", ""},
	{"jež alfred na zadatku", "jez alfred na zadatku", "", "medium", ""},
	{"kalimero", "", "kalimero - ", "medium", "kalimero- crtani"},
	{"kasper", "", "", "medium", "kasper crtani"},
	{"konanove avanture", "", "", "long", ""},
	{"kuče dragoljupče", "kuce dragoljupce", "", "medium", ""},
	{"lale gator", "", "", "medium", ""},
	{"la linea", "", "", "medium", ""},
	{"legenda o tarzanu", "", "", "long", ""},
	{"le piaf", "", "", "short", ""},
	{"liga super zloća", "liga super zloca", "", "medium", ""},
	{"mali detektivi", "", "", "long", ""},
	{"mali leteći medvjedići", "mali leteci medvjedici", "", "long", ""},
	{"masa i medved", "masha i medved", "masa i medvjed", "medium", "masa i medved crtani"},
	{"mačor mika", "macor mika", "", "long", ""},
	{"mece dobrići", "mece dobrici", "", "medium", ""},
	{"miki maus", "", "", "medium", ""},
	{"mornar popaj", "", "", "medium", ""},
	{"mr. bean", "mr bean", "mr.bean", "medium", "mr bean animated"},
	{"mumijevi", "", "", "medium", ""},
	{"nindža kornjače", "nindza kornjace", "ninja kornjace", "long", ""},
	{"ogi i žohari", "ogi i zohari", "", "long", ""},
	{"otkrića bez granica", "otkrica bez granica", "", "long", ""},
	{"paja patak", "", "", "medium", ""},
	{"patak dača", "patak daca", "", "medium", ""},
	{"pepa prase", "", "", "medium", ""},
	{"pepe le tvor", "", "", "medium", ""},
	{"pera detlić", "pera detlic", "", "medium", ""},
	{"pera kojot", "", "", "medium", ""},
	{"pingvini sa madagaskara", "", "", "medium", ""},
	{"pink panter", "", "", "medium", "pink panter crtani"},
	{"plava princeza", "", "", "long", ""},
	{"porodica kremenko", "", "", "long", ""},
	{"poručnik draguljče", "porucnik draguljce", "", "medium", ""},
	{"princeze sirene", "", "", "long", ""},
	{"profesor baltazar", "", "", "medium", ""},
	{"ptica trkačica", "ptica trkacica", "", "medium", ""},
	{"pustolovine sa braćom kret", "pustolovine sa bracom kret", "", "long", ""},
	{"rakuni", "", "", "long", ""},
	{"ratnik kišna kap", "ratnik kisna kap", "", "long", ""},
	{"ren i stimpi", "", "", "medium", ""},
	{"robotek", "", "robotech", "long", ""},
	{"šalabajzerići", "salabajzerici", "", "medium", ""},
	{"silvester", "", "silvester i tviti", "medium", "silvester crtani"},
	{"šilja", "silja", "", "medium", "silja crtani"},
	{"snorkijevci", "", "", "medium", ""},
	{"sofronije", "", "", "medium", ""},
	{"super miš", "super mis", "", "medium", "super mis crtani"},
	{"supermen", "", "", "medium", "supermen crtani"},
	{"super špijunke", "super spijunke", "", "long", ""},
	{"sport bili", "", "", "medium", ""},
	{"srle i pajče", "srle i pajce", "", "medium", ""},
	{"stanlio i olio", "", "", "medium", ""},
	{"stari crtaći", "stari crtaci", "stari sinhronizovani crtaci", "medium", ""},
	{"stripi", "", "", "medium", ""},
	{"štrumfovi", "strumpfovi", "strumfovi", "medium", "strumfovi crtani"},
	{"sundjer bob kockalone", "sundjer bob", "sunđer bob", "medium", ""},
	{"talični tom", "talicni tom", "", "long", ""},
	{"tarzan gospodar džungle", "tarzan gospodar dzungle", "", "long", ""},
	{"tom i džeri", "tom i dzeri", "", "medium", ""},
	{"transformersi", "", "", "long", ""},
	{"vitez koja", "", "", "medium", ""},
	{"voltron force", "", "", "long", "voltron force crtani"},
	{"vuk vučko", "vuk vucko", "", "medium", ""},
	{"wumi", "", "wummi", "short", "wumi crtani"},
	{"zamenik boža", "zamenik boza", "", "medium", ""},
	{"zemlja konja", "", "", "medium", ""},
	{"zmajeva kugla", "zmajeva kugla", "zmajeva kugla z", "long", ""},
}

var filters = []string{
	"najbolji crtaci",
	"www.crtani-filmovi.org",
	"by crtani svijet",
	"crtanifilmonline",
	"crtani filmovi",
	"crtani film",
	"stari crtani",
	"cijeli crtani",
	"crtani",
	"crtic",
	"sinhronizovano",
	"sihronizovano",
	"sinhronizovani",
	"sinhronizovan",
	"sinhronizacija",
	"sinkronizacija",
	"titlovano",
	"sa prevodom",
	"nove epizode",
	"na srpskom jeziku",
	"na srpskom",
	"srpska",
	"srpski",
	"srb ",
	" srb",
	" sd",
	" hq",
	" svcd",
	"hrvatska",
	"hrv,srp,bos",
	"zagorci",
	"slovenska verzija",
	"b92 tv",
	"za decu",
	"zadecu",
	"youtube",
	"youtube",
	"full movie",
	"mashini skazki",
	"the cartooner 100",
	"iz 60-70-80-tih",
	"mpeg4",
	"144p h 264 aac",
	"sihroni fll 2",
	"zlekedik",
	"gusztav allast keres",
	"guszt v k",
	"rtb",
	"tvrip",
	"djuza stoiljkovic",
	"okrenite preko smplayer-a",
	"new episodes",
	"new episode",
	"animado",
	"animad8",
	"animated",
	"cartoon",
	"of 47",
	"dailymotion video",
	"video dailymotion",
	"ultra tv",
	"happy tv",
}

var censoredWords = []string{
	"kurac",
	"kurcu",
	"sranje",
	"sranja",
	"govna",
	"govno",
	"picka",
	"picke",
	"peder",
	"jebač",
	"uzivo",
	"parodija",
	"tretmen",
	"ispaljotka",
	"kinder jaja",
	"video igrice",
	"atomski mravi",
	"igracke od plastelina",
	"sex",
	"sexy",
	"flesh",
	"ubisoft",
	"wanna",
	"special",
	"trailer",
	"teaser",
	"music",
	"monster",
	"intro",
	"countdown",
	"eternity",
	"summer",
	"galaxy",
	"constitution",
	"hunkyard",
	"riders",
	"flash",
	"wanted",
	"instrumental",
	"gamer",
	"remix",
	"tour",
	"party",
	"bjorke",
	"tweety",
	"revolution",
	"halloween",
	"remastered",
	"celebration",
	"experiments",
	"food",
	"gameplay",
	"surprise",
	"batters",
	"bottle",
	"erasers",
	"series",
	"comics",
	"village",
	"theatre",
	"dolphin",
	"stallone",
	"koniec",
	"latino",
	"lovers",
	"lubochka",
	"tobbe",
	"sushi",
	"prikljuchenija",
	"slovenska",
	"aakerhus",
	"sylvia",
	"deutsch",
	"remue",
	"kespar",
	"splitter",
	"desierto",
	"pelicula",
	"episodio",
	"rwerk",
	"xhaven",
	"erkste",
	"przytul",
	"potenzf",
	"szalony",
	"schweiz",
	"verkackt",
	"sottile",
	"goldene",
	"osterhase",
	"elasmosaurio",
	"ombra",
	"ehlers",
	"dejas",
	"capitulo",
	"et ses",
	"tu sais",
	"ma vision",
	"v riti ",
	"how could",
	"new year",
	" del ",
	"maldicion",
	"fernsehausstrahlung",
	"the best bits of",
	"jamella",
	"kasper-sky",
	"kasper internet",
	"feiern",
	"terpidana",
	"borba za koralovo",
	"sve pesme",
	"minecraft",
	"gasttoz",
	"gastoz",
	"batailles",
	"maminka",
}

var censoredIds = []string{
	"52vfFeJERfQ",
	"DLk4SLmIDUU",
	"VsNOHQfm02M",
	"yfYfGnCVbHs",
	"8zZ6tg2LXiM",
	"-1CnR5qVh5E",
	"DPxb3-7lakw",
	"zKhPpVTUn_Q",
	"vBggIcqV1rc",
	"YrmKYtDnthk",
	"YzqWmqeR43I",
	"Id3kHQC9vPI",
	"Ngke-HPnHok",
	"7VPtdqnHxHw",
	"Q6hTJ11ZGwU",
	"n28-lRu5cpw",
	"dl_kPk276oo",
	"YsdOt6qc6o4",
	"Tm7mOlgPlxA",
	"X8BwFSHJpg4",
	"QYrsrjgGh5g",
	"_z6pgpPDXBY",
	"7_ys2vKapLg",
	"G7SnbTCsj28",
	"2LzVPEoiacY",
	"8QJozzsvPnU",
	"KpLrIWB78sQ",
	"CuV0mDu4GL4",
	"c1-ywGJfS8U",
	"sotlkpiczWk",
	"wPUhMP7aGnw",
	"AR0Jc1rh2N0",
	"xuGex-B3GbQ",
	"drEJEbHDgIA",
	"JF4qkkgQsO4",
	"Y4r7m-Payv8",
	"0HDbPXN-HaE",
	"yONB3IwxtlQ",
	"BjtFDLOmEu8",
	"cpF73znG7UM",
	"2onSjJVgtpg",
	"gRLppbNNCLI",
	"smKnRR0ouds",
	"5nElGb8odmk",
	"xs_IiToWEEs",
	"co4B3-BwcUY",
	"QCHSP32z2nc",
	"CRyDcmPHZy4",
	"YyFNnHlDgP0",
	"uVn9vpFlljE",
	"LuYPcHVsyow",
	"6sPavaRoEA8",
	"JpG72eCdmKo",
	"XL4TZVlGc3o",
	"bAeHz5miEJg",
	"6zCKejl_1bY",
	"Ragr70eHvQg",
	"7nDvQPYJpMg",
	"jYU_doi3Y7w",
	"5FqM9elA3AY",
	"3Wha_dlJ9G0",
	"bTLVcWEtk-0",
	"WIqCOdLvDBs",
	"x5-SgQVUY-c",
	"MQPLlgJkvLc",
	"Fu88sn45nlE",
	"0Uq52kdn-MA",
	"9sMBmQNPFTA",
	"AiUBMthSJ9I",
	"jAsHYIyp9fY",
	"8nWgj3tk0kg",
	"2lD_oXT4ssA",
	"iT4ZXYso2kA",
	"fuYMMWpRjkM",
	"xy53o1",
	"xy53q1",
	"x3osiz",
	"x25ja2c",
	"x4aha4",
	"x7k5hx",
	"x60rr7",
	"x7k5ko",
	"x4e7mn",
	"xs4jyr",
	"x7wviw",
	"x5nyjf",
	"x20uqv5",
	"x20uvfy",
	"x29i8ae",
	"x2k0tnk",
	"x2etl16",
	"x2dv1ec",
	"x196m5x",
	"x2cz5es",
	"4562474",
	"21508130",
	"14072389",
	"145041047",
	"163980203",
	"165572645",
	"168855693",
	"61534934",
	"15376700",
	"73551241",
	"80060489",
}

var chain = `-----BEGIN CERTIFICATE-----
MIIDfTCCAuagAwIBAgIDErvmMA0GCSqGSIb3DQEBBQUAME4xCzAJBgNVBAYTAlVT
MRAwDgYDVQQKEwdFcXVpZmF4MS0wKwYDVQQLEyRFcXVpZmF4IFNlY3VyZSBDZXJ0
aWZpY2F0ZSBBdXRob3JpdHkwHhcNMDIwNTIxMDQwMDAwWhcNMTgwODIxMDQwMDAw
WjBCMQswCQYDVQQGEwJVUzEWMBQGA1UEChMNR2VvVHJ1c3QgSW5jLjEbMBkGA1UE
AxMSR2VvVHJ1c3QgR2xvYmFsIENBMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
CgKCAQEA2swYYzD99BcjGlZ+W988bDjkcbd4kdS8odhM+KhDtgPpTSEHCIjaWC9m
OSm9BXiLnTjoBbdqfnGk5sRgprDvgOSJKA+eJdbtg/OtppHHmMlCGDUUna2YRpIu
T8rxh0PBFpVXLVDviS2Aelet8u5fa9IAjbkU+BQVNdnARqN7csiRv8lVK83Qlz6c
JmTM386DGXHKTubU1XupGc1V3sjs0l44U+VcT4wt/lAjNvxm5suOpDkZALeVAjmR
Cw7+OC7RHQWa9k0+bw8HHa8sHo9gOeL6NlMTOdReJivbPagUvTLrGAMoUgRx5asz
PeE4uwc2hGKceeoWMPRfwCvocWvk+QIDAQABo4HwMIHtMB8GA1UdIwQYMBaAFEjm
aPkr0rKV10fYIyAQTzOYkJ/UMB0GA1UdDgQWBBTAephojYn7qwVkDBF9qn1luMrM
TjAPBgNVHRMBAf8EBTADAQH/MA4GA1UdDwEB/wQEAwIBBjA6BgNVHR8EMzAxMC+g
LaArhilodHRwOi8vY3JsLmdlb3RydXN0LmNvbS9jcmxzL3NlY3VyZWNhLmNybDBO
BgNVHSAERzBFMEMGBFUdIAAwOzA5BggrBgEFBQcCARYtaHR0cHM6Ly93d3cuZ2Vv
dHJ1c3QuY29tL3Jlc291cmNlcy9yZXBvc2l0b3J5MA0GCSqGSIb3DQEBBQUAA4GB
AHbhEm5OSxYShjAGsoEIz/AIx8dxfmbuwu3UOx//8PDITtZDOLC5MH0Y0FWDomrL
NhGc6Ehmo21/uBPUR/6LWlxz/K7ZGzIZOKuXNBSqltLroxwUCEm2u+WR74M26x1W
b8ravHNjkOR/ez4iyz0H7V84dJzjA1BOoa+Y7mHyhD8S
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIEKjCCAxKgAwIBAgIEOGPe+DANBgkqhkiG9w0BAQUFADCBtDEUMBIGA1UEChML
RW50cnVzdC5uZXQxQDA+BgNVBAsUN3d3dy5lbnRydXN0Lm5ldC9DUFNfMjA0OCBp
bmNvcnAuIGJ5IHJlZi4gKGxpbWl0cyBsaWFiLikxJTAjBgNVBAsTHChjKSAxOTk5
IEVudHJ1c3QubmV0IExpbWl0ZWQxMzAxBgNVBAMTKkVudHJ1c3QubmV0IENlcnRp
ZmljYXRpb24gQXV0aG9yaXR5ICgyMDQ4KTAeFw05OTEyMjQxNzUwNTFaFw0yOTA3
MjQxNDE1MTJaMIG0MRQwEgYDVQQKEwtFbnRydXN0Lm5ldDFAMD4GA1UECxQ3d3d3
LmVudHJ1c3QubmV0L0NQU18yMDQ4IGluY29ycC4gYnkgcmVmLiAobGltaXRzIGxp
YWIuKTElMCMGA1UECxMcKGMpIDE5OTkgRW50cnVzdC5uZXQgTGltaXRlZDEzMDEG
A1UEAxMqRW50cnVzdC5uZXQgQ2VydGlmaWNhdGlvbiBBdXRob3JpdHkgKDIwNDgp
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArU1LqRKGsuqjIAcVFmQq
K0vRvwtKTY7tgHalZ7d4QMBzQshowNtTK91euHaYNZOLGp18EzoOH1u3Hs/lJBQe
sYGpjX24zGtLA/ECDNyrpUAkAH90lKGdCCmziAv1h3edVc3kw37XamSrhRSGlVuX
MlBvPci6Zgzj/L24ScF2iUkZ/cCovYmjZy/Gn7xxGWC4LeksyZB2ZnuU4q941mVT
XTzWnLLPKQP5L6RQstRIzgUyVYr9smRMDuSYB3Xbf9+5CFVghTAp+XtIpGmG4zU/
HoZdenoVve8AjhUiVBcAkCaTvA5JaJG/+EfTnZVCwQ5N328mz8MYIWJmQ3DW1cAH
4QIDAQABo0IwQDAOBgNVHQ8BAf8EBAMCAQYwDwYDVR0TAQH/BAUwAwEB/zAdBgNV
HQ4EFgQUVeSB0RGAvtiJuQijMfmhJAkWuXAwDQYJKoZIhvcNAQEFBQADggEBADub
j1abMOdTmXx6eadNl9cZlZD7Bh/KM3xGY4+WZiT6QBshJ8rmcnPyT/4xmf3IDExo
U8aAghOY+rat2l098c5u9hURlIIM7j+VrxGrD9cv3h8Dj1csHsm7mhpElesYT6Yf
zX1XEC+bBAlahLVu2B064dae0Wx5XnkcFMXj0EyTO2U87d89vqbllRrDtRnDvV5b
u/8j72gZyxKTJ1wDLW8w0B62GqzeWvfRqqgnpv55gcR5mTNXuhKwqeBCbJPKVt7+
bYQLCIt+jerXmCHG8+c8eS9enNFMFY3h7CI3zJpDC5fcgJCNs2ebb0gIFVbPv/Er
fF6adulZkMV8gzURZVE=
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIElDCCA3ygAwIBAgIQAf2j627KdciIQ4tyS8+8kTANBgkqhkiG9w0BAQsFADBh
MQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkwFwYDVQQLExB3
d3cuZGlnaWNlcnQuY29tMSAwHgYDVQQDExdEaWdpQ2VydCBHbG9iYWwgUm9vdCBD
QTAeFw0xMzAzMDgxMjAwMDBaFw0yMzAzMDgxMjAwMDBaME0xCzAJBgNVBAYTAlVT
MRUwEwYDVQQKEwxEaWdpQ2VydCBJbmMxJzAlBgNVBAMTHkRpZ2lDZXJ0IFNIQTIg
U2VjdXJlIFNlcnZlciBDQTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEB
ANyuWJBNwcQwFZA1W248ghX1LFy949v/cUP6ZCWA1O4Yok3wZtAKc24RmDYXZK83
nf36QYSvx6+M/hpzTc8zl5CilodTgyu5pnVILR1WN3vaMTIa16yrBvSqXUu3R0bd
KpPDkC55gIDvEwRqFDu1m5K+wgdlTvza/P96rtxcflUxDOg5B6TXvi/TC2rSsd9f
/ld0Uzs1gN2ujkSYs58O09rg1/RrKatEp0tYhG2SS4HD2nOLEpdIkARFdRrdNzGX
kujNVA075ME/OV4uuPNcfhCOhkEAjUVmR7ChZc6gqikJTvOX6+guqw9ypzAO+sf0
/RR3w6RbKFfCs/mC/bdFWJsCAwEAAaOCAVowggFWMBIGA1UdEwEB/wQIMAYBAf8C
AQAwDgYDVR0PAQH/BAQDAgGGMDQGCCsGAQUFBwEBBCgwJjAkBggrBgEFBQcwAYYY
aHR0cDovL29jc3AuZGlnaWNlcnQuY29tMHsGA1UdHwR0MHIwN6A1oDOGMWh0dHA6
Ly9jcmwzLmRpZ2ljZXJ0LmNvbS9EaWdpQ2VydEdsb2JhbFJvb3RDQS5jcmwwN6A1
oDOGMWh0dHA6Ly9jcmw0LmRpZ2ljZXJ0LmNvbS9EaWdpQ2VydEdsb2JhbFJvb3RD
QS5jcmwwPQYDVR0gBDYwNDAyBgRVHSAAMCowKAYIKwYBBQUHAgEWHGh0dHBzOi8v
d3d3LmRpZ2ljZXJ0LmNvbS9DUFMwHQYDVR0OBBYEFA+AYRyCMWHVLyjnjUY4tCzh
xtniMB8GA1UdIwQYMBaAFAPeUDVW0Uy7ZvCj4hsbw5eyPdFVMA0GCSqGSIb3DQEB
CwUAA4IBAQAjPt9L0jFCpbZ+QlwaRMxp0Wi0XUvgBCFsS+JtzLHgl4+mUwnNqipl
5TlPHoOlblyYoiQm5vuh7ZPHLgLGTUq/sELfeNqzqPlt/yGFUzZgTHbO7Djc1lGA
8MXW5dRNJ2Srm8c+cftIl7gzbckTB+6WohsYFfZcTEDts8Ls/3HB40f/1LkAtDdC
2iDJ6m6K7hQGrn2iWZiIqBtvLfTyyRRfJs8sjX7tN8Cp1Tm5gr8ZDOo0rwAhaPit
c+LJMto4JQtV05od8GiG7S5BNO98pVAdvzr508EIDObtHopYJeS4d60tbvVS3bR0
j6tJLp07kzQoH3jOlOrHvdPJbRzeXDLz
-----END CERTIFICATE-----`

var (
	reTitle = regexp.MustCompile(`[0-9A-Za-zžćčšđ_,]+`)
	reAlpha = regexp.MustCompile(`[A-Za-zžćčšđ]+`)
	reDesc  = regexp.MustCompile(`(?U)(\(|\[).*(\)|\])`)
	reYear  = regexp.MustCompile(`(19\d{2}|20\d{2})`)
	reExt   = regexp.MustCompile(`\.(?i:avi|mp4|flv|wmv|mpg|mpeg|mpeg4)$`)
	reRip   = regexp.MustCompile(`(?i:xvid)?(tv|dvd)?(-|\s)(rip)(bg)?(audio)?`)
	reChars = regexp.MustCompile(`(?i:braca grimm|i snupi [sš]ou|i snupi|charlie brown and snoopy|brzi gonzales i patak da[cč]a|patak da[cč]a i brzi gonzales|patak da[cč]a i elmer|patak da[cč]a i gicko prasi[cć]|i hello kitty|tom and jerry|tom i d[zž]eri [sš]ou|spongebob squarepants|paja patak i [sš]ilja|bini i sesil|masha i medved|elmer fudd|blinkibil|kockalone|najlepse bajke|stari sinhronizovani crtaci|popeye the sailor|kasper i drugari,|leghorn)`)
	reTime  = regexp.MustCompile(`(\d{2})h(\d{2})m(\d{2})s`)
	rePart  = regexp.MustCompile(`\s([\diI]{1,2})\.?\s?(?i:/|deo|od|part)\s?([\diI]{1,2})?\s*(?i:deo)?`)
	rePart2 = regexp.MustCompile(`\s(?i:pt)\s?(\d{1,2})\s*`)

	reTitleR     = regexp.MustCompile(`^(\d{1,2}\.?)\s?(\d{1,})?(.*)$`)
	reTitleNotEp = regexp.MustCompile(`\d{2,}\s(?i:razbojnika|sati|malih|pljeskavica)`)
	reTitle20    = regexp.MustCompile(`(\s20)`)

	reE1 = regexp.MustCompile(`(?i:epizoda|epizida|epzioda|episode|epizodas|episoda|Эпизод)\s?(\d{1,3})`)
	reE2 = regexp.MustCompile(`(\d{1,3})\.?-?\s?(?i:epizoda|epizida|epzioda|episode|epizodas|episoda)`)
	reE3 = regexp.MustCompile(`\s(?i:ep|e)\.?\s*(\d{1,3})`)
	reE4 = regexp.MustCompile(`(?:^|-|\.|\s)\s?(\d{1,3}\b)`)
	reE5 = regexp.MustCompile(`(?i:s)(?:\d{1,2})(?i:e)(\d{2})(?:\d{1})?(?:a|b)`)
	reE6 = regexp.MustCompile(`(?i:s)(?:\d{1,2})(?:e)(\d{1,2})`)

	reS1 = regexp.MustCompile(`(?i:sezona|sezon)\s?(\d{1,2})`)
	reS2 = regexp.MustCompile(`(\d{1,2})\.?\s?(?i:sezona|sezon)`)
	reS3 = regexp.MustCompile(`(?i:s)\s?(\d{1,2})`)
	reS4 = regexp.MustCompile(`(\d{1,2})(?i:x)`)
)

var (
	wg       sync.WaitGroup
	cartoons []cartoon
)

type multiSorter struct {
	cartoons []cartoon
}

func (ms *multiSorter) Sort(cartoons []cartoon) {
	ms.cartoons = cartoons
	sort.Sort(ms)
}

func (ms *multiSorter) Len() int {
	return len(ms.cartoons)
}

func (ms *multiSorter) Swap(i, j int) {
	ms.cartoons[i], ms.cartoons[j] = ms.cartoons[j], ms.cartoons[i]
}

func (ms *multiSorter) Less(i, j int) bool {
	p, q := &ms.cartoons[i], &ms.cartoons[j]

	episode := func(c1, c2 *cartoon) bool {
		if c1.Episode == -1 {
			return false
		} else if c2.Episode == -1 {
			return true
		}
		return c1.Episode < c2.Episode
	}

	season := func(c1, c2 *cartoon) bool {
		if c1.Season == -1 {
			return false
		} else if c2.Season == -1 {
			return true
		}
		return c1.Season < c2.Season
	}

	switch {
	case season(p, q):
		return true
	case season(q, p):
		return false
	case episode(p, q):
		return true
	case episode(q, p):
		return false
	}

	return episode(p, q)
}

func decodePem(certInput string) tls.Certificate {
	var cert tls.Certificate
	certPEMBlock := []byte(certInput)
	var certDERBlock *pem.Block
	for {
		certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)
		if certDERBlock == nil {
			break
		}
		if certDERBlock.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, certDERBlock.Bytes)
		}
	}
	return cert
}

func getTLSConfig() tls.Config {
	certChain := decodePem(chain)
	conf := tls.Config{}
	conf.RootCAs = x509.NewCertPool()
	for _, cert := range certChain.Certificate {
		x509Cert, err := x509.ParseCertificate(cert)
		if err != nil {
			panic(err)
		}
		conf.RootCAs.AddCert(x509Cert)
	}
	conf.BuildNameToCertificate()
	return conf
}

func youTube(char character) {

	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			log.Print("Recovered in youTube:", r)
		}
	}()

	const apiKey = "YOUR_API_KEY"

	tlsConfig := getTLSConfig()
	tr := http.Transport{TLSClientConfig: &tlsConfig}

	httpClient := &http.Client{
		Transport: &transport.APIKey{Key: apiKey, Transport: &tr},
	}

	yt, err := youtube.New(httpClient)
	if err != nil {
		log.Print("Error creating YouTube client:", err)
		return
	}

	name := strings.ToLower(char.Name)
	altname := strings.ToLower(char.AltName)
	altname2 := strings.ToLower(char.AltName2)

	getResponse := func(token string) *youtube.SearchListResponse {
		apiCall := yt.Search.List("id,snippet").
			Q(getQuery(char, false)).
			MaxResults(50).
			VideoDuration(char.Duration).
			Type("video").
			PageToken(token)

		response, err := apiCall.Do()
		if err != nil {
			log.Print("Error making YouTube API call:", err.Error())
			return nil
		}
		return response
	}

	parseResponse := func(response *youtube.SearchListResponse) {
		for _, video := range response.Items {
			videoId := video.Id.VideoId
			videoTitle := strings.ToLower(video.Snippet.Title)
			videoThumbSmall := video.Snippet.Thumbnails.Default.Url
			videoThumbMedium := video.Snippet.Thumbnails.Medium.Url
			videoThumbLarge := video.Snippet.Thumbnails.High.Url

			if isValidTitle(videoTitle, name, altname, altname2, videoId) {
				formattedTitle := getFormattedTitle(videoTitle, name, altname, altname2)

				c := cartoon{
					videoId,
					name,
					videoTitle,
					formattedTitle,
					getEpisode(videoTitle),
					getSeason(videoTitle),
					"youtube",
					"https://www.youtube.com/watch?v=" + videoId,
					videoThumbSmall,
					videoThumbMedium,
					videoThumbLarge,
				}

				cartoons = append(cartoons, c)
			}
		}
	}

	response := getResponse("")
	parseResponse(response)

	if response.NextPageToken != "" {
		response = getResponse(response.NextPageToken)
		parseResponse(response)
	}

}

func dailyMotion(char character) {

	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			log.Print("Recovered in dailyMotion:", r)
		}
	}()

	uri := "https://api.dailymotion.com/videos?search=%s&fields=id,title,url,duration,thumbnail_120_url,thumbnail_360_url,thumbnail_480_url&limit=50&page=%s&sort=relevance"

	name := strings.ToLower(char.Name)
	altname := strings.ToLower(char.AltName)
	altname2 := strings.ToLower(char.AltName2)

	timeout := time.Duration(6 * time.Second)

	dialTimeout := func(network, addr string) (net.Conn, error) {
		return net.DialTimeout(network, addr, timeout)
	}

	tlsConfig := getTLSConfig()

	transport := http.Transport{
		Dial:            dialTimeout,
		TLSClientConfig: &tlsConfig,
	}

	httpClient := http.Client{
		Transport: &transport,
	}

	getResponse := func(page string) ([]interface{}, bool) {
		res, err := httpClient.Get(fmt.Sprintf(uri, getQuery(char, true), page))
		if err != nil {
			log.Print("Error making DailyMotion API call:", err.Error())
			return nil, false
		}
		body, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()

		var data map[string]interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			log.Print("Error unmarshaling json:", err.Error())
			return nil, false
		}

		hasMore, ok := data["has_more"].(bool)
		if !ok {
			return nil, false
		}

		response, ok := data["list"].([]interface{})
		if !ok {
			return nil, false
		}

		if len(response) == 0 {
			return nil, false
		}

		return response, hasMore
	}

	parseResponse := func(response []interface{}) {
		for _, obj := range response {
			video, ok := obj.(map[string]interface{})
			if !ok {
				continue
			}

			videoId := video["id"].(string)
			videoTitle := strings.ToLower(video["title"].(string))
			videoUrl := video["url"].(string)
			videoThumbSmall := video["thumbnail_120_url"].(string)
			videoThumbMedium := video["thumbnail_360_url"].(string)
			videoThumbLarge := video["thumbnail_480_url"].(string)

			videoDuration := getDuration(video["duration"].(float64))

			if isValidTitle(videoTitle, name, altname, altname2, videoId) && char.Duration == videoDuration {
				formattedTitle := getFormattedTitle(videoTitle, name, altname, altname2)

				c := cartoon{
					videoId,
					name,
					videoTitle,
					formattedTitle,
					getEpisode(videoTitle),
					getSeason(videoTitle),
					"dailymotion",
					videoUrl,
					videoThumbSmall,
					videoThumbMedium,
					videoThumbLarge,
				}

				cartoons = append(cartoons, c)
			}
		}
	}

	response, hasMore := getResponse("1")
	if response != nil {
		parseResponse(response)
	}

	if hasMore {
		response, _ := getResponse("2")
		if response != nil {
			parseResponse(response)
		}
	}

}

func vimeo(char character) {

	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			log.Print("Recovered in vimeo:", r)
		}
	}()

	const apiKey = "YOUR_API_KEY"
	uri := "https://api.vimeo.com/videos?query=%s&page=%s&per_page=100&sort=relevant"

	name := strings.ToLower(char.Name)
	altname := strings.ToLower(char.AltName)
	altname2 := strings.ToLower(char.AltName2)

	timeout := time.Duration(6 * time.Second)

	dialTimeout := func(network, addr string) (net.Conn, error) {
		return net.DialTimeout(network, addr, timeout)
	}

	tlsConfig := getTLSConfig()

	transport := http.Transport{
		Dial:            dialTimeout,
		TLSClientConfig: &tlsConfig,
	}

	httpClient := http.Client{
		Transport: &transport,
	}

	getResponse := func(page string) []interface{} {
		req, err := http.NewRequest("GET", fmt.Sprintf(uri, getQuery(char, true), page), nil)
		if err != nil {
			log.Print("Error making Vimeo API call: %v", err.Error())
			return nil
		}

		req.Header.Set("Authorization", "bearer "+apiKey)
		req.Header.Set("Accept", "application/vnd.vimeo.video+json;version=3.2")
		res, err := httpClient.Do(req)
		if err != nil {
			log.Print("Error making Vimeo API call:", err.Error())
			return nil
		}
		body, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()

		var data map[string]interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			log.Print("Error unmarshaling json:", err.Error())
			return nil
		}

		response, ok := data["data"].([]interface{})
		if !ok {
			return nil
		}

		if len(response) == 0 {
			return nil
		}

		return response
	}

	parseResponse := func(response []interface{}) {
		for _, obj := range response {
			video, ok := obj.(map[string]interface{})
			if !ok {
				continue
			}

			videoId := strings.Replace(video["link"].(string), "https://vimeo.com/", "", -1)
			videoTitle := strings.ToLower(video["name"].(string))
			videoUrl := video["link"].(string)

			pictures, ok := video["pictures"].(map[string]interface{})
			if !ok {
				continue
			}

			sizes := pictures["sizes"].([]interface{})

			if len(sizes) < 4 {
				continue
			}

			videoThumbSmall := sizes[3].(map[string]interface{})["link"].(string)
			videoThumbMedium := sizes[2].(map[string]interface{})["link"].(string)
			videoThumbLarge := sizes[1].(map[string]interface{})["link"].(string)

			videoDuration := getDuration(video["duration"].(float64))

			if isValidTitle(videoTitle, name, altname, altname2, videoId) && char.Duration == videoDuration {
				formattedTitle := getFormattedTitle(videoTitle, name, altname, altname2)

				c := cartoon{
					videoId,
					name,
					videoTitle,
					formattedTitle,
					getEpisode(videoTitle),
					getSeason(videoTitle),
					"vimeo",
					videoUrl,
					videoThumbSmall,
					videoThumbMedium,
					videoThumbLarge,
				}

				cartoons = append(cartoons, c)
			}
		}
	}

	response := getResponse("1")
	if response != nil {
		parseResponse(response)
	}

}

func getDuration(videoDuration float64) string {
	minutes := videoDuration / 60
	switch {
	case minutes < 4 && minutes > 0:
		return "short"
	case minutes >= 4 && minutes <= 20:
		return "medium"
	case minutes > 20:
		return "long"
	default:
		return "any"
	}
}

func getFormattedTitle(videoTitle string, name string, altname string, altname2 string) string {

	title := videoTitle

	part := ""
	p := rePart.FindAllStringSubmatch(title, -1)
	if len(p) > 0 {
		part = p[0][1]
	}

	p2 := rePart2.FindAllStringSubmatch(title, -1)
	if len(p2) > 0 {
		part = p2[0][1]
	}

	title = reYear.ReplaceAllString(title, "")

	re20 := reTitle20.FindAllStringSubmatch(title, -1)
	if len(re20) > 1 {
		title = reTitle20.ReplaceAllString(title, " ")
	}

	for _, filter := range filters {
		title = strings.Replace(title, filter, " ", -1)
	}

	for _, re := range []*regexp.Regexp{
		reDesc, reExt, reRip, reChars, reTime, rePart, rePart2,
		reE1, reS1, reS4, reE2, reE5, reE6, reE3, reS2, reS3} {
		title = re.ReplaceAllString(title, "")
	}

	matches := reTitle.FindAllString(title, -1)
	title = strings.Join(matches, " ")
	title = strings.Replace(title, "_", " ", -1)

	name = strings.Replace(name, "-", "", -1)
	name = strings.TrimRight(name, " ")

	if altname2 != "" {
		title = strings.Replace(title, altname2, "", 1)
	}
	if altname != "" {
		title = strings.Replace(title, altname+" ", "", 1)
	}
	title = strings.Replace(title, name, "", 1)

	title = strings.TrimLeft(title, " ")
	title = strings.TrimRight(title, " ")

	title = reTitleR.ReplaceAllString(title, "$3")

	if strings.HasPrefix(title, "i ") || strings.HasPrefix(title, "and ") || strings.HasPrefix(title, " i ") {
		title = fmt.Sprintf("%s %s", name, title)
	}

	if !reAlpha.MatchString(title) {
		title = name
	}

	if part != "" {
		title = fmt.Sprintf("%s - %s deo", title, part)
	}

	return title
}

func getEpisode(videoTitle string) int {
	title := videoTitle

	title = reYear.ReplaceAllString(title, "")

	re20 := reTitle20.FindAllStringSubmatch(title, -1)
	if len(re20) > 1 {
		title = reTitle20.ReplaceAllString(title, " ")
	}

	for _, filter := range filters {
		title = strings.Replace(title, filter, " ", -1)
	}
	for _, re := range []*regexp.Regexp{reDesc, reYear, reTime, rePart, rePart2, reS1, reS4} {
		title = re.ReplaceAllString(title, "")
	}

	ep := -1
	e1 := reE1.FindAllStringSubmatch(title, -1)
	if len(e1) > 0 {
		ep, _ = strconv.Atoi(e1[0][1])
		return ep
	}

	e2 := reE2.FindAllStringSubmatch(title, -1)
	if len(e2) > 0 {
		ep, _ = strconv.Atoi(e2[0][1])
		return ep
	}

	e5 := reE5.FindAllStringSubmatch(title, -1)
	if len(e5) > 0 {
		ep, _ = strconv.Atoi(e5[0][1])
		return ep
	}

	e6 := reE6.FindAllStringSubmatch(title, -1)
	if len(e6) > 0 {
		ep, _ = strconv.Atoi(e6[0][1])
		return ep
	}

	e3 := reE3.FindAllStringSubmatch(title, -1)
	if len(e3) > 0 {
		ep, _ = strconv.Atoi(e3[0][1])
		return ep
	}

	e4 := reE4.FindAllStringSubmatch(title, -1)
	notEp := reTitleNotEp.MatchString(title)
	if len(e4) > 0 && !notEp {
		ep, _ = strconv.Atoi(e4[0][1])
		if ep > 100 || ep == 0 {
			return -1
		}
		return ep
	}

	return ep
}

func getSeason(videoTitle string) int {
	title := videoTitle

	title = reYear.ReplaceAllString(title, "")

	re20 := reTitle20.FindAllStringSubmatch(title, -1)
	if len(re20) > 1 {
		title = reTitle20.ReplaceAllString(title, " ")
	}

	for _, re := range []*regexp.Regexp{reDesc, reYear, reTime, rePart, rePart2, reE1} {
		title = re.ReplaceAllString(title, "")
	}

	s := -1
	s1 := reS1.FindAllStringSubmatch(title, -1)
	if len(s1) > 0 {
		s, _ = strconv.Atoi(s1[0][1])
		return s
	}

	s2 := reS2.FindAllStringSubmatch(title, -1)
	if len(s2) > 0 {
		s, _ = strconv.Atoi(s2[0][1])
		return s
	}

	s3 := reS3.FindAllStringSubmatch(title, -1)
	if len(s3) > 0 {
		s, _ = strconv.Atoi(s3[0][1])
		if s >= 20 || s == 0 {
			return -1
		}
		return s
	}

	s4 := reS4.FindAllStringSubmatch(title, -1)
	if len(s4) > 0 {
		s, _ = strconv.Atoi(s4[0][1])
		if s >= 20 || s == 0 {
			return -1
		}
		return s
	}

	return s
}

func getQuery(char character, escape bool) string {
	query := ""
	if char.Query != "" {
		query = char.Query
	} else if char.AltName != "" {
		query = char.AltName
	} else {
		query = char.Name
	}
	if escape {
		query = url.QueryEscape(query)
	}
	return query
}

func isCensored(videoTitle string, videoId string) bool {
	for _, word := range censoredWords {
		if strings.Contains(videoTitle, word) {
			return true
		}
	}
	for _, id := range censoredIds {
		if id == videoId {
			return true
		}
	}
	return false
}

func isValidTitle(videoTitle string, name string, altname string, altname2 string, videoId string) bool {
	videoTitle = reTitleR.ReplaceAllString(videoTitle, "$3")
	videoTitle = strings.TrimLeft(videoTitle, " ")

	if strings.HasPrefix(videoTitle, name) {
		if !isCensored(videoTitle, videoId) {
			return true
		}
	}
	if altname != "" {
		if strings.HasPrefix(videoTitle, altname) {
			if !isCensored(videoTitle, videoId) {
				return true
			}
		}
	}
	if altname2 != "" {
		if strings.HasPrefix(videoTitle, altname2) {
			if !isCensored(videoTitle, videoId) {
				return true
			}
		}
	}
	return false
}

func List() (string, error) {
	js, err := json.MarshalIndent(characters, "", "    ")
	if err != nil {
		return "empty", err
	}
	return string(js[:]), nil
}

func Search(query string) (string, error) {
	char := new(character)
	for _, c := range characters {
		if query == c.Name || query == c.AltName {
			char = &c
			break
		}
	}

	if char.Name != "" {
		wg.Add(3)
		cartoons = make([]cartoon, 0)
		go youTube(*char)
		go dailyMotion(*char)
		go vimeo(*char)
		wg.Wait()

		ms := multiSorter{}
		ms.Sort(cartoons)

		js, err := json.MarshalIndent(cartoons, "", "    ")
		if err != nil {
			return "empty", err
		}

		return string(js[:]), nil
	} else {
		return "empty", nil
	}
}

func Extract(service string, videoId string) (string, error) {
	var url string
	var err error
	switch {
	case service == "youtube":
		url, err = vidextr.YouTube(videoId)
	case service == "dailymotion":
		url, err = vidextr.DailyMotion(videoId)
	case service == "vimeo":
		url, err = vidextr.Vimeo(videoId)
	}

	if err != nil {
		return "empty", err
	}

	if url == "" {
		return "empty", nil
	}

	js, err := json.Marshal(url)
	if err != nil {
		return "empty", err
	}
	return string(js[:]), nil
}

func ListenAndServe(bind string) {
	http.HandleFunc("/list", handleList)
	http.HandleFunc("/search", handleSearch)
	http.HandleFunc("/extract", handleExtract)

	l, err := net.Listen("tcp4", bind)
	if err != nil {
		log.Fatal(err)
	}
	http.Serve(l, nil)
}

func setHeader(w http.ResponseWriter) {
	w.Header().Set("Server", fmt.Sprintf("%s/%s", appName, appVersion))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func handleList(w http.ResponseWriter, r *http.Request) {
	setHeader(w)
	js, err := List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(js))
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	setHeader(w)

	query := r.FormValue("q")

	if query != "" {
		js, err := Search(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if js == "" {
			http.Error(w, "404 Not Found", http.StatusNotFound)
			return
		}
		w.Write([]byte(js))
	} else {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}
}

func handleExtract(w http.ResponseWriter, r *http.Request) {
	setHeader(w)

	service := r.FormValue("srv")
	videoId := r.FormValue("id")

	if service != "" && videoId != "" {
		js, err := Extract(service, videoId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if js == "empty" {
			http.Error(w, "", http.StatusNotFound)
			return
		} else {
			w.Write([]byte(js))
			return
		}
	} else {
		http.Error(w, "", http.StatusForbidden)
		return
	}
}

func main() {
	bind := flag.String("bind", ":7313", "Bind address")
	flag.Parse()
	ListenAndServe(*bind)
}