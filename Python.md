## Hvað er Python?

Python er túlkað, æðra forritunarmál (high-level language) og sérstaklega þekkt fyrir læsileika og einfaldleika miðað við mörg önnur forritunarmál. Python er gríðarlega vinsælt ekki síst vegna þess hversu sveigjnlegt það er og allra þeirra skráarsafna sem það getur notað og búið er að þróa í kringum það og hjálpa til við lausn ýmissa verkefna. Hugbúnaðarhönnuðir um heim allan nota Python til ýmissa verka eins og í vefforritun, gervigreind, vísindalegri forritun, gagnavísindum o.fl. Python er einnig vel nýtilegt sem reiknivél og hjálpartæki við ýmis verkefni í námi. Python er svokallað hlutbundið forritunarmál.

### Af hverju Python?

* Auðvelt að læra: Einföld málfræði og læsileiki gera Python sérlega heppilegt fyrir byrjendur í forritun.

* Sveigjanleiki: Notað mjög víða við ýmis verkefni

* Samfélag og leiðbeiningar: Mjög stórt samfélag er til í kringum Python á vefnum og mikið til af hjálpargögnum og leiðbeiningum.

### Fyrsta Python forritið og samanburður við C

Hér er dæmi um Python forrit sem prentar Halló heimur út á skjáinn.

```Python
print("Halló heimur")
```
Til samanburðar væri sambærilegt forrit í forritunarmálinu C eftirfarandi:

```C
#include<stdio.h>
int main()
{
    printf("Halló heimur");
    return 0;
}
```
C er reyndar annað eldra og mjög öflugt forritunarmál en Python og C tengjast nánum böndum því Python málið er skrifað í C.

C er orðið um 50 ára gamalt og var upphaflega þróað samhliða Unix stýrikerfinu sem var mikil bylting á tölvusviðinu og Unix stýrikerfið er allt skrifað í C. C er einkar hentugt mál til að skrifa stýrikerfi og forrita örgjörva og smátölvur því það er það forritunarmál sem næst kemst vélbúnaðinum fyrir utan smalamál (assembly) sem er mun stirðara og óþægilegra í notkun. C er þýtt (compiled) forritunarmál og gríðarlega hraðskreitt, en Python er túlkað (interpreted) og mun hægvirkara. Við munum síðar kíkja á C (og systurmál þess C++) samhliða skoðun á Arduino örtölvunni en hér munum við fyrst um sinn einbeita okkur að Python.

**Æfing 1**
Notaðu Python til að skrifa nafnið þitt með print() skipuninni.

### Breyturskilgreiningar og breytugerðir

Breytur eru geymslusvæði (í minni tölvunnar) fyrir gögn sem unnið er með. Python inniheldur nokkrar breytugerðir eins og heiltölubreytur (integer), fleytitölubreytur (float), strengbreytur (string) og rökbreytur (boolean).

```Python
aldur = 30 # heiltala
haed = 1.80 # fleytitala
nafn = "Anna" # textastrengur
frost_uti = True # rökbreyta (satt eða ósatt)
```

**Æfing 2**
Búðu til breytur sem geyma upplýsingar um aldur þinn, nafnið þitt og upplýsingar um hvort það sé satt eða ósatt að þér finnist Python skemmtilegt forritunarmál.

### Reikniaðgerðir strengjaaðferðir

Python styður reikniaðgerðirnar fjórar (+, -, *, /) og kann einnig að meðhöndla strengbreytur.

```Python
# Reikniaðgerðir
samlagning = 3 + 7
fradrattur = 10 - 3
margfoldun = 4 * 5
deiling = 20 / 4

# Strengjaaðgerðir
kvedja = "Halló" + " " + "heimur" 
hlatur = "Ha" * 3
```
**Æfing 3**
Skrifaðu tvö Python forrit sem 
a) reiknar flatarmál ferhyrnings
b) reiknar flatarmál rafmagnsvírs

### Flæðistjórnun og ákvörðunartaka

`if`, `elif` og `else` lykilorðin eru innbyggð í Python og hjálpa til við að taka ákvarðanir í forritum. 

> **ATHUGIÐ**
Inndráttur setninga er mjög mikilvægur í Python og málið er mjög strangt á því að réttum inndrætti sé fylgt. Ef það er ekki gert þá kvartar túlkurinn og flytur villuskilaboð. Flest nýrri umhverfi og textaritlar sjá að mestu um þennan inndrátt. Margir telja þetta einn helsta kostinn við Python að málið neyðir forritarann til að skrifa læsileg forrit. Aðrir telja þetta mjög hamlandi. Hvað sem því líður þá þarf að gæta vel að því að inndrátturinn sé réttur</u> 

```Python
x = 10
if x > 5:
    print("x er stærra en 5")
elif x == 5:
    print("x er jafnt og 5")
else:
    print("x er minna en 5")
```
**Æfing 4**
Skrifaðu forrit sem athugar hvort að tala sé jákvæð, neikvæð eða núll.

### Lykkjur

Lykkjur eru eitt af því sem gerir forritun og tölvur jafn öflugar og raun ber vitni, eða hæfileiki tölva til að endurtaka aðgerðir mjög hratt. Python hefur yfir að ráða tveimur gerðum af lykkjum `for` og `while`. 

```Python
# For lykkjan
for i in range(5):
    print(i)

# While lykkjan
teljari = 0
while teljari < 5:
    print(count)
    count += 1
```
**Æfing 5**
Skrifaðu forrit sem að prentar út 7-sinnum töfluna með því að nota lykkju.

### Föll
Föll eru forritsblokkir sem eiga að framkvæma ákveðið verk. Þegar þú þarft vinna ákveðið verk geturðu skrifað fall sem þú getur svo alltaf kallað í aftur of aftur í forritinu.

```Python
# Fallið skilgreint
def heilsa():
    print("Komdu sæll Python forritari!")

# Kallað í fallið
heilsa()
```
**Æfing 6**
Búðu til fall sem að tekur inn nafn sem færibreytu (parameter) og prentar svo út kveðju.

**Stærra verkefni**
Búðu til Python forrit sem að umbreytir hitastigi úr selsíus kvarðanum og yfir í fahrenheit og öfugt

$C = (F-32)\times \frac{5}{9}$
$F = (C \times \frac{9}{5}+32$