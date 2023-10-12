## Veftækni. Stutt kynning á notkun HTML við gerð vefsíðu.

Í þessu verkefni ætlum við að búa til vinna með vefinn sjálfann og skoða hvað er á bakvið vefsíður í grunninn og skoða einnig vinnslu með hreinum texta í textaritli (text editor).

Allskonar tól og tæki eru til í dag til að einfalda vefsíðugerð og til heilu vefumsjónarkerfin. En á bakvið vefsíður er alltaf í grunninn skriftarmálið HTML sem lesið er af staðlinum HTTP (Hypertext Transfer Protocol) sem er innbyggður í vefskoðara. En hér ætlum við að gera þetta handvirkt og skrifa HTML beint í textaritil, en áður en fínni tól komu til sögunnar voru fyrstu vefsíðurnar allar smíðaðar þannig.

Þegar vefur er gerður þá er nauðsynlegt og allra mikilvægast af öllu að hafa efni til að koma á framfæri. Hér ráðið þið efnisvali, þið gætuð gert kynningu á ykkur sjálfum og helstu áhugamálum, nú eða fundið annað efni ef þið hafið aðrar hugmyndir.

## 1. Textaritill

Finnið textaritil (text editor) til að vinna með. Textaritlar vinna með hreinan texta, en tekið skal fram að forrit eins og Word gera það ekki. Afar margir ritlar eru til og margir hverjir eru opin og frír hugbúnaður. Hér er listi yfir nokkra þekkta og mikið notaða textaritla.

* **<a href="https://code.visualstudio.com/">Visual Studio Code</a>** (Windows, MacOS, Linux)
    - Afar vinsæll og mjög fullkominn / Opin hugbúnaður
* **<a href="https://notepad-plus-plus.org/">Notepad++</a>** (Windows)
    - Mjög fullkominn og endurbættur Notepad / Opin hugbúnaður
* **TextEdit** (MacOS) 
    - Staðalbúnaður í MacOS / Lokaður hugbúnaður
* **Notepad** (Windows)
    - Staðalbúnaður í Windows / Lokaður hugbúnaður

## 2. Vefsíða

Þegar þú hefur fundið textaritil við hæfi þá skaltu gera eftirfarandi:

1. Búðu til möppu, t.d. á skjáborðinu sem að heitir t.d. **WWW** eða **Vefur**.

2. Opnaðu textaritilinn og búðu til nýja skrá í möppunni þinni sem heitir `index.html` og vistaðu hana í möppunni.

```html
<!DOCTYPE html>
<html>
<head>
    <title>Vefsíðan mín</title>
</head>
<body>

</body>
</html>
```
Þetta er HTML skriftarmálið sem þú ert að nota en það byggir á tögum (tags) sem segja til um hvernig eigi að forsníða textann á síðunni. Yfirleitt opnast tög `<>` og svo er þeim lokað `</>` Allar vefsíður verða að hafa `<html>`, `<head>` og `<body>` tögin. 

3. Inn í `<body>` tagið skaltu bæta við `<h1>` (heading 1) taginu og skrifa þar t.d. nafnið þitt:

```html

<h1>Nafnið þitt hér</h1>

```

Vistaðu skrána.

4. Síðan skaltu nota `<p>` (paragraph) tagið til þess að til að búa til málsgrein:

```html
<p>Hér er ég að búa til feykilega spennandi vefsíðu og mig hlakkar til þegar hún verður tilbúin.</p>
```

Vistaðu skrána.

5. Setjum inn lista yfir áhugamál. Til þess notum við `<ul>` og `<li>`

```html
<h2>Áhugamálin mín:</h2>
<ul>
    <li>Lesa góðar bækur</li>
    <li>Sofa út á morgnanna</li>
    <li>Fara í sund</li>
    <li>Ferðast</li>
</ul>
```

Vistaðu skrána

6. Bætum nú inn mynd. Finnið einhverja góða mynd og hlaðið henni niður af vefnum inn í möppuna ykkar (athugið að gæta þarf að höfundarétti þegar myndir eru teknar af vefnum). Notum ``<img>`` tagið til þess að setja myndina inn í vefsíðuna.

```html
<img src="nafn-a-mynd.jpg" alt="Lýsing á mynd">
```

Einnig má setja inn tengil á mynd á vefnum þannig

```html
<img src="https://www.fnv.is/static/themes/2016/images/logo.png?v=6a1bec472a">
```

Vistaðu skrána.

7. Nú er lokaprófun eftir. Opnaðu forritið í vefskoðara til þess að sjá hvernig hún myndi líta út á vefnum. Athugaðu hvort allt sé eins og þú vilt hafa það.

8. Farðu á vefsíðuna <a href="https://www.w3schools.com/">**w3schools**</a> og skoðaðu þig um. Þessi síða er afar gagnleg til að læra allskonar veftækni. Svaraðu eftirfarandi spurningum með hjálp síðunnar.

- Hvað er CSS?

- Hvað er Javascript?

9. Reyndu nú að breyta, bæta við og setja inn eitthvað fleira. Notaðu hugmyndaflugið. Kíktu á síðu eins og w3schools á vefnum til að læra meira, fikta og fá hugmyndir. Prófaðu t.d. að setja inn 

* Töflu
* Tengil í aðra vefsíðu eða tengil á mynd
* Breyttu lit á texta (þar kemur inn sextándakerfið)
* o.fl. sem þér dettur í hug.

**Skilaðu endanlegri textaskrá `index.html` í skilakassann á Moodle**
