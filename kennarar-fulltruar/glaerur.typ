// ============================================================
//  Gervigreindarfulltrúar í skólastarfi
//  Glærur fyrir kennara — um 30–40 mínútur
//
//  Venjuleg útgáfa:      typst compile glaerur.typ
//  Með glósum kynnis:    typst compile glaerur.typ --input notes=true glaerur-glosur.pdf
// ============================================================

#let show-notes = sys.inputs.at("notes", default: "false") == "true"

// ---------- Litir ----------
#let navy        = rgb("#1E3A5F")
#let chat-c      = rgb("#2F6FD6")
#let chat-soft   = rgb("#E7EFFC")
#let agent-c     = rgb("#0E8A6A")
#let agent-soft  = rgb("#E1F4EE")
#let accent      = rgb("#E8891A")
#let accent-soft = rgb("#FDF1E3")
#let red         = rgb("#C0392B")
#let red-soft    = rgb("#FBE9E7")
#let gray        = rgb("#6B7280")
#let light       = rgb("#F2F4F7")
#let ink         = rgb("#1F2937")

// ---------- Útlit ----------
#set document(title: "Gervigreindarfulltrúar í skólastarfi")
#set page(
  width: 29.7cm,
  height: if show-notes { 19.7cm } else { 16.71cm },
  margin: (x: 1.6cm, top: 1.2cm, bottom: if show-notes { 4.2cm } else { 1.2cm }),
  footer: context {
    let n = counter(page).get().first()
    if n > 1 and not show-notes {
      set text(size: 11pt, fill: gray)
      [Gervigreindarfulltrúar í skólastarfi]
      h(1fr)
      str(n)
    }
  },
)
#set text(
  font: ("Inter", "Noto Sans", "Liberation Sans", "DejaVu Sans"),
  size: 22pt, fill: ink, lang: "is", hyphenate: false,
)
#set par(leading: 0.6em)
#set list(marker: text(fill: accent, size: 0.8em)[●], spacing: 0.75em)
#set enum(spacing: 0.75em)

// ---------- Hjálparföll ----------
#let notes(body) = if show-notes {
  place(bottom + left, dy: 3.4cm, block(
    width: 100%, fill: rgb("#FFF9E6"), stroke: 1pt + accent, inset: 7pt, radius: 4pt,
    text(size: 12pt, fill: ink)[*Glósur:* #body],
  ))
}

#let slide(title: none, body) = {
  pagebreak(weak: true)
  if title != none {
    text(size: 30pt, weight: "bold", fill: navy, title)
    v(-0.5em)
    box(width: 2.6cm, height: 5pt, fill: accent, radius: 3pt)
    v(0.5em)
  }
  body
}

#let section(num, title, sub, note: none) = page(fill: navy, footer: none)[
  #set text(fill: white)
  #align(horizon)[
    #text(size: 18pt, fill: accent, weight: "bold")[HLUTI #num]
    #v(0.1em)
    #text(size: 46pt, weight: "bold", title)
    #v(0.1em)
    #text(size: 22pt, fill: rgb("#C9D6E8"), sub)
  ]
  #if note != none { notes(note) }
]

#let arrow-r(c: gray, len: 1.4cm, label: none) = align(center + horizon, stack(
  dir: ttb, spacing: 6pt,
  ..if label != none { (text(size: 14pt, fill: c, style: "italic", label),) },
  box(width: len, height: 16pt, {
    place(horizon + left, line(length: len - 10pt, stroke: 3.5pt + c))
    place(horizon + right, polygon(fill: c, (0pt, -8pt), (13pt, 0pt), (0pt, 8pt)))
  }),
))

#let arrow-d(c: gray, len: 0.9cm) = align(center, box(width: 16pt, height: len, {
  place(center + top, line(angle: 90deg, length: len - 10pt, stroke: 3.5pt + c))
  place(center + bottom, polygon(fill: c, (-8pt, 0pt), (0pt, 13pt), (8pt, 0pt)))
}))

#let actor(label, fill: navy, r: 1.2cm, size: 20pt) = box(circle(
  radius: r, fill: fill, inset: 0pt,
  align(center + horizon, text(fill: white, weight: "bold", size: size, label)),
))

#let node(body, fill: light, stroke: none, w: auto, size: 19pt, color: ink) = box(
  fill: fill, stroke: stroke, radius: 12pt, inset: (x: 14pt, y: 12pt), width: w,
  align(center, text(size: size, fill: color, body)),
)

#let flow(..items) = grid(
  columns: items.pos().len(), column-gutter: 8pt, align: center + horizon, ..items.pos(),
)

#let card(title: none, fill: light, color: navy, height: auto, body) = block(
  width: 100%, height: height, fill: fill, radius: 14pt, inset: 16pt, {
    if title != none {
      text(size: 21pt, weight: "bold", fill: color, title)
      v(-0.2em)
    }
    text(size: 18pt, body)
  },
)

#let tag(body, color: navy) = text(size: 14pt, weight: "bold", fill: color, tracking: 1pt, body)

#let says(body) = block(fill: chat-soft, radius: 14pt, inset: 18pt, width: 100%)[
  #tag(color: chat-c)[ÞÚ SEGIR]
  #v(-0.2em)
  #text(size: 20pt, style: "italic", body)
]

#let does(..items) = block(fill: agent-soft, radius: 14pt, inset: 18pt, width: 100%)[
  #tag(color: agent-c)[FULLTRÚINN …]
  #v(-0.1em)
  #set text(size: 19pt)
  #for it in items.pos() [
    #grid(columns: (auto, 1fr), column-gutter: 10pt, text(fill: agent-c, weight: "bold")[✓], it)
    #v(-0.15em)
  ]
]

#let big-quote(body, color: navy) = align(center + horizon, block(width: 88%,
  text(size: 30pt, weight: "bold", fill: color, body)))

// ============================================================
//  1. Titilsíða
// ============================================================
#page(fill: navy, footer: none)[
  #set text(fill: white)
  #align(horizon)[
    #text(size: 18pt, fill: accent, weight: "bold", tracking: 1.5pt)[GERVIGREIND Í SKÓLASTARFI]
    #v(0.2em)
    #text(size: 50pt, weight: "bold")[Frá spjalli \ til verka]
    #v(0.3em)
    #text(size: 24pt, fill: rgb("#C9D6E8"))[Hvað eru gervigreindarfulltrúar \ — og hvernig geta kennarar nýtt þá?]
  ]
  #place(right + horizon, dx: -0.4cm, stack(dir: ltr, spacing: 14pt,
    actor("Spjall", fill: chat-c, r: 1.7cm, size: 18pt),
    align(horizon, text(size: 40pt, fill: accent)[→]),
    actor("Fulltrúi", fill: agent-c, r: 1.7cm, size: 18pt),
  ))
  #notes[Bjóða fólk velkomið. Segja að engin tækniþekking sé nauðsynleg — í dag tölum við um hugmyndir, ekki tækni. (1 mín.)]
]

// ============================================================
//  2. Dagskrá
// ============================================================
#slide(title: "Dagskrá")[
  #v(0.4em)
  #let item(n, t, c) = grid(columns: (1.4cm, 1fr), column-gutter: 14pt, align: horizon,
    actor(str(n), fill: c, r: 0.55cm, size: 18pt), text(size: 24pt, t))
  #stack(spacing: 18pt,
    item(1, [Spjall — það sem við þekkjum], chat-c),
    item(2, [Fulltrúi — hvað er nýtt?], agent-c),
    item(3, [Fulltrúar í skólastofunni], accent),
    item(4, [Að vinna vel og örugglega], navy),
    item(5, [Umræður og spurningar], gray),
  )
  #notes[Fara stuttlega yfir. Spurningar eru velkomnar hvenær sem er. (1 mín.)]
]

// ============================================================
//  3. Upphitun
// ============================================================
#slide(title: "Réttið upp hönd …")[
  #v(1em)
  #grid(columns: (1fr, 1fr, 1fr), column-gutter: 22pt,
    card(fill: chat-soft, color: chat-c, title: "1")[
      #text(size: 22pt)[… sem hafa *prófað* gervigreindarspjall, t.d. ChatGPT, Copilot, Gemini eða Claude.]],
    card(fill: accent-soft, color: accent, title: "2")[
      #text(size: 22pt)[… sem nota það *í hverri viku* í starfi.]],
    card(fill: agent-soft, color: agent-c, title: "3")[
      #text(size: 22pt)[… sem hafa látið gervigreind *vinna verk* fyrir sig, ekki bara svara.]],
  )
  #v(1fr)
  #align(center, text(size: 20pt, fill: gray)[Flest okkar þekkja spjallið. Í dag skoðum við næsta skref.])
  #notes[Láta fólk rétta upp hönd fyrir hverja spurningu. Gefur tilfinningu fyrir hópnum. Líklega fáir við nr. 3 — það er í góðu lagi! (2 mín.)]
]

// ============================================================
//  HLUTI 1 — Spjall
// ============================================================
#section(1, "Spjall", "Það sem flest okkar þekkja")

#slide(title: "Svona virkar spjall")[
  #v(1.2em)
  #align(center, flow(
    actor("Þú"),
    arrow-r(label: "spurning", c: chat-c),
    node(fill: chat-soft, size: 22pt, color: chat-c, w: 4.2cm)[*Spjall*],
    arrow-r(label: "svar", c: chat-c),
    actor("Þú"),
    arrow-r(c: gray),
    node(fill: accent-soft, size: 19pt, w: 5.6cm)[*Þú* gerir \ restina sjálf/ur],
  ))
  #v(1fr)
  #align(center, text(size: 28pt, weight: "bold", fill: navy)[Þú spyrð. Það svarar. \ Svo er boltinn hjá þér.])
  #v(1fr)
  #notes[Spjall er eins og samtal: ein spurning, eitt svar. Gervigreindin gerir ekkert fyrr en þú skrifar næst, og hún gerir ekkert utan spjallgluggans. (2 mín.)]
]

#slide(title: "Spjall er frábært til að …")[
  #v(0.6em)
  #grid(columns: (1fr, 1fr), column-gutter: 22pt, row-gutter: 18pt,
    card(fill: chat-soft, color: chat-c, title: "Fá hugmyndir")[„Gefðu mér 10 hugmyndir að upphafi kennslustundar um bergtegundir.“],
    card(fill: chat-soft, color: chat-c, title: "Útskýra")[„Útskýrðu brot fyrir 9 ára barni.“],
    card(fill: chat-soft, color: chat-c, title: "Skrifa drög")[„Skrifaðu drög að tölvupósti til foreldra um vettvangsferð.“],
    card(fill: chat-soft, color: chat-c, title: "Svara spurningum")[„Hver er munurinn á lýsingarorði og atviksorði?“],
  )
  #v(1fr)
  #align(center, text(size: 24pt, weight: "bold", fill: red)[… en spjallið *gerir* ekkert sjálft.])
  #notes[Hrósa spjallinu — það er gagnlegt! En það endar alltaf á texta á skjá sem þú þarft sjálf/ur að vinna úr. (2 mín.)]
]

#slide(title: "Afrita – líma – laga – endurtaka")[
  #let step(t) = node(fill: light, size: 19pt, w: 6.2cm, t)
  #v(0.3em)
  #grid(columns: (1fr, auto, 1fr), align: center + horizon, row-gutter: 6pt, column-gutter: 10pt,
    step[Spyrja spjallið], arrow-r(), step[Afrita svarið],
    [], [], arrow-d(),
    step[Laga útlit og villur], box(rotate(180deg, arrow-r())), step[Líma inn í skjal],
    arrow-d(), [], [],
    step[Spyrja aftur …], [], [],
  )
  #place(right + bottom, dy: -0.3cm, block(width: 11.5cm, fill: accent-soft, radius: 14pt, inset: 16pt)[
    #text(size: 22pt, weight: "bold", fill: accent)[Þú ert sendillinn.]
    #v(-0.3em)
    #text(size: 18pt)[Gervigreindin hugsar — en *þú* berð allt á milli staða og vinnur handavinnuna.]
  ])
  #notes[Kannast einhver við þetta? Spyrja salinn. Þetta er punkturinn þar sem fulltrúar breyta leiknum. (2 mín.)]
]

// ============================================================
//  HLUTI 2 — Fulltrúi
// ============================================================
#section(2, "Fulltrúi", "Gervigreind sem vinnur verkið")

#slide(title: "Hvað er fulltrúi?")[
  #v(1fr)
  #align(center, block(width: 90%, fill: agent-soft, radius: 18pt, inset: 30pt)[
    #set text(size: 28pt)
    Fulltrúi #text(fill: gray, size: 20pt)[(e. _AI agent_)] er gervigreind sem fær \
    *#text(fill: agent-c)[markmið]* — ekki bara spurningu — \
    og *vinnur sjálf í mörgum skrefum* \
    þar til verkinu er lokið.
  ])
  #v(1fr)
  #align(center, grid(columns: 2, column-gutter: 50pt,
    text(fill: chat-c, size: 24pt)[Spjall: *„Hvað ætti ég að gera?“*],
    text(fill: agent-c, size: 24pt)[Fulltrúi: *„Gerðu það.“*],
  ))
  #v(1fr)
  #notes[Lykilskilgreiningin. Orðið „fulltrúi“ — einhver sem kemur fram fyrir þína hönd og sinnir erindum fyrir þig. (2 mín.)]
]

#slide(title: "Svona vinnur fulltrúi")[
  #v(0.8em)
  #align(center, flow(
    actor("Þú"),
    arrow-r(label: "markmið", c: agent-c),
    block(stroke: (paint: agent-c, thickness: 2.5pt, dash: "dashed"), radius: 18pt, inset: 14pt)[
      #flow(
        node(fill: agent-soft, size: 18pt)[*Skipuleggur*],
        arrow-r(len: 0.9cm, c: agent-c),
        node(fill: agent-soft, size: 18pt)[*Framkvæmir*],
        arrow-r(len: 0.9cm, c: agent-c),
        node(fill: agent-soft, size: 18pt)[*Athugar*],
      )
      #v(-0.1em)
      #text(size: 16pt, fill: agent-c, style: "italic")[↻ endurtekur þar til verkinu er lokið]
    ],
    arrow-r(label: "afrakstur", c: agent-c),
    actor("Þú"),
  ))
  #v(1fr)
  #grid(columns: (1fr, 1fr, 1fr), column-gutter: 18pt,
    card(title: "Skipuleggur", color: agent-c)[Brýtur verkið niður í skref.],
    card(title: "Framkvæmir", color: agent-c)[Opnar skjöl, leitar, skrifar, vistar.],
    card(title: "Athugar", color: agent-c)[Skoðar eigin vinnu og lagar.],
  )
  #notes[Mikilvægasta myndin. Fulltrúinn vinnur í hring — skipuleggur, framkvæmir, athugar — og kemur svo til baka með tilbúið verk. Þú ferð yfir í lokin. (3 mín.)]
]

#slide(title: "Leiðarlýsing eða leigubíll?")[
  #v(0.5em)
  #grid(columns: (1fr, 1fr), column-gutter: 28pt,
    block(fill: chat-soft, radius: 16pt, inset: 22pt, width: 100%, height: 9.2cm)[
      #tag(color: chat-c)[SPJALL] \
      #text(size: 30pt, weight: "bold", fill: chat-c)[Leiðarlýsing]
      #v(0.2em)
      #text(size: 19pt, style: "italic")[„Beygðu til vinstri, svo til hægri, svo beint áfram …“]
      #v(0.4em)
      Þú færð góð ráð — \ en *þú keyrir sjálf/ur*.
    ],
    block(fill: agent-soft, radius: 16pt, inset: 22pt, width: 100%, height: 9.2cm)[
      #tag(color: agent-c)[FULLTRÚI] \
      #text(size: 30pt, weight: "bold", fill: agent-c)[Leigubíll]
      #v(0.2em)
      #text(size: 19pt, style: "italic")[„Ég þarf að komast í Hörpu.“]
      #v(0.4em)
      Bílstjórinn finnur leiðina \ og *keyrir þig þangað*. \ Þú segir hvert — og fylgist með.
    ],
  )
  #notes[Myndlíking sem flestir skilja strax. Með leigubíl þarftu samt að segja skýrt hvert þú ætlar — það sama á við um fulltrúa. (2 mín.)]
]

#slide(title: "Eða í skólamáli …")[
  #v(0.5em)
  #grid(columns: (1fr, 1fr), column-gutter: 28pt,
    card(fill: chat-soft, color: chat-c, height: 8.6cm, title: "Spjall = samkennari á kaffistofunni")[
      #v(0.3em)
      #text(size: 21pt)[Gefur þér góð ráð og hugmyndir.]
      #v(0.3em)
      #text(size: 21pt)[Svo ferð þú aftur í stofuna og *vinnur verkið sjálf/ur*.]
    ],
    card(fill: agent-soft, color: agent-c, height: 8.6cm, title: "Fulltrúi = aðstoðarmaður")[
      #v(0.3em)
      #text(size: 21pt)[Þú felur honum verkefni og hann *vinnur það*.]
      #v(0.3em)
      #text(size: 21pt)[Hann kemur til baka með afraksturinn — og *þú ferð yfir* áður en það fer lengra.]
    ],
  )
  #notes[Tengja við daglegan veruleika kennara: stuðningsfulltrúi eða kennaranemi sem fær verkefni og skilar því. (1–2 mín.)]
]

#slide(title: "Munurinn í hnotskurn")[
  #v(0.2em)
  #set text(size: 20pt)
  #table(
    columns: (1.2fr, 1fr, 1fr),
    inset: (x: 14pt, y: 11pt),
    stroke: none,
    fill: (x, y) => if y == 0 { none } else if x == 1 { chat-soft } else if x == 2 { agent-soft } else { light },
    align: (x, y) => if x == 0 { left + horizon } else { center + horizon },
    [], text(fill: chat-c, weight: "bold", size: 24pt)[Spjall], text(fill: agent-c, weight: "bold", size: 24pt)[Fulltrúi],
    [*Hvað gefur þú?*], [Spurningu], [Markmið],
    [*Hver vinnur verkið?*], [Þú], [Fulltrúinn],
    [*Fjöldi skrefa*], [Eitt svar í einu], [Mörg skref í röð],
    [*Getur opnað skjöl og forrit?*], [Takmarkað], [Já — með þínu leyfi],
    [*Hvað færðu?*], [Texta á skjá], [Tilbúið verk, t.d. skjal],
  )
  #notes[Samantekt á muninum. Ekki lesa allt upp — benda á línuna „Hver vinnur verkið?“. (2 mín.)]
]

#slide(title: "Sama verkefni — tvær leiðir")[
  #align(center, block(fill: accent-soft, radius: 12pt, inset: 12pt, width: 100%,
    text(size: 17pt)[*Verkefni:* Lesskilningsverkefni um eldgos fyrir 6. bekk, á þremur erfiðleikastigum, með svörum.]))
  #v(0.3em)
  #let step(n, t, c) = grid(columns: (0.8cm, 1fr), column-gutter: 6pt, align: horizon,
    actor(str(n), fill: c, r: 0.3cm, size: 12pt), text(size: 16pt, t))
  #grid(columns: (1fr, 1fr), column-gutter: 24pt,
    block(fill: chat-soft, radius: 14pt, inset: 16pt, width: 100%)[
      #tag(color: chat-c)[MEÐ SPJALLI]
      #v(-0.1em)
      #stack(spacing: 4pt,
        step(1, [Biðja um texta], chat-c),
        step(2, [Afrita yfir í Word], chat-c),
        step(3, [Biðja um spurningar], chat-c),
        step(4, [Afrita aftur og raða], chat-c),
        step(5, [Biðja um léttari og þyngri útgáfu], chat-c),
        step(6, [Afrita, líma, laga útlit …], chat-c),
        step(7, [Búa til svarblöð], chat-c),
      )
    ],
    block(fill: agent-soft, radius: 14pt, inset: 16pt, width: 100%)[
      #tag(color: agent-c)[MEÐ FULLTRÚA]
      #v(-0.1em)
      #stack(spacing: 4pt,
        step(1, [Þú lýsir verkefninu *einu sinni*], agent-c),
        step(2, [Fulltrúinn vinnur á meðan þú sinnir öðru], agent-c),
        step(3, [Þú ferð yfir *þrjú tilbúin skjöl*], agent-c),
      )
      #v(1fr)
      #align(center, text(size: 22pt, weight: "bold", fill: agent-c)[Þú: 2 skref í stað 7])
    ],
  )
  #notes[Raunhæft dæmi. Tíminn sem sparast er handavinnan — afrita, líma, laga útlit. Faglega vinnan (yfirlesturinn) er enn þín. (2–3 mín.)]
]

#slide(title: "Verkfærakista fulltrúans")[
  #v(0.4em)
  #let tool(name, ex, c, soft) = block(fill: soft, radius: 14pt, inset: 16pt, width: 100%, height: 3.6cm)[
    #text(size: 22pt, weight: "bold", fill: c, name)
    #v(-0.3em)
    #text(size: 16pt, fill: ink, ex)
  ]
  #grid(columns: (1fr, 1fr, 1fr), column-gutter: 18pt, row-gutter: 18pt,
    tool("Skjöl og glærur", [Býr til, les og lagar Word-skjöl og glærur.], agent-c, agent-soft),
    tool("Vefurinn", [Leitar að upplýsingum og efni.], chat-c, chat-soft),
    tool("Töflureiknar", [Raðar einkunnum, listum og yfirlitum.], accent, accent-soft),
    tool("Möppur og skrár", [Finnur, flokkar og vistar gögn.], agent-c, agent-soft),
    tool("Tölvupóstur og dagatal", [Skrifar drög og finnur tíma.], chat-c, chat-soft),
    tool("… og fleira", [Fer eftir því hvaða kerfi skólinn notar.], gray, light),
  )
  #v(1fr)
  #align(center, text(size: 20pt)[Fulltrúinn notar *sömu verkfæri og þú* — en aðeins þau sem *þú leyfir*.])
  #notes[Fulltrúar hafa hendur: þeir geta opnað og vistað skrár. Nefna dæmi um þjónustur sem bjóða fulltrúa (t.d. Claude, ChatGPT, Microsoft Copilot, Google Gemini) — en það breytist hratt, og fylgja skal því sem skólinn hefur samþykkt. (2 mín.)]
]

// ============================================================
//  HLUTI 3 — Í skólastofunni
// ============================================================
#section(3, "Í skólastofunni", "Hvernig geta kennarar nýtt fulltrúa?")

#slide(title: "Fjögur svið þar sem fulltrúar hjálpa")[
  #v(0.4em)
  #grid(columns: (1fr, 1fr), column-gutter: 22pt, row-gutter: 20pt,
    card(fill: agent-soft, color: agent-c, title: "1 · Undirbúningur")[Kennsluáætlanir, verkefni, glærur og próf.],
    card(fill: chat-soft, color: chat-c, title: "2 · Aðlögun")[Mismunandi erfiðleikastig, einfalt mál, þýðingar.],
    card(fill: accent-soft, color: accent, title: "3 · Endurgjöf")[Drög að endurgjöf út frá matskvarðanum þínum.],
    card(fill: light, color: navy, title: "4 · Skipulag og samskipti")[Fréttabréf, tölvupóstar og röð og regla á gögnum.],
  )
  #notes[Yfirlit. Næstu fjórar glærur sýna eitt dæmi fyrir hvert svið: hvað þú segir og hvað fulltrúinn gerir. (1 mín.)]
]

#slide(title: "1 · Undirbúningur kennslu")[
  #v(0.4em)
  #grid(columns: (1.15fr, 1fr), column-gutter: 24pt,
    says[„Búðu til kennsluáætlun til fjögurra vikna um vatnið í náttúrunni fyrir 4. bekk, í samræmi við aðalnámskrá. Eitt verklegt verkefni á viku. Vistaðu sem skjal í möppunni Náttúrufræði.“],
    does(
      [Les hæfniviðmið í aðalnámskrá],
      [Skiptir efninu á vikurnar],
      [Semur verkefni og efnislista],
      [Setur upp og vistar skjalið],
    ),
  )
  #v(1fr)
  #align(center, text(size: 20pt, fill: gray)[Þú færð *fyrstu drög* — og lagar þau að þínum nemendum.])
  #notes[Benda á að fyrirmælin eru eins og verkbeiðni: hvað, fyrir hverja, hvernig og hvar á að vista. (2 mín.)]
]

#slide(title: "2 · Aðlögun að ólíkum nemendum")[
  #v(0.4em)
  #grid(columns: (1.15fr, 1fr), column-gutter: 24pt,
    says[„Taktu þennan texta og gerðu þrjár útgáfur: létta, miðlungs og krefjandi. Gerðu líka útgáfu á einfaldri íslensku með orðalista og aðra á pólsku.“],
    does(
      [Endurskrifar textann á þremur stigum],
      [Einfaldar mál og býr til orðalista],
      [Þýðir yfir á pólsku],
      [Setur allt snyrtilega í eina möppu],
    ),
  )
  #v(1fr)
  #align(center, text(size: 20pt, fill: gray)[Það sem áður tók kvöld tekur nú *nokkrar mínútur* að yfirfara.])
  #notes[Einstaklingsmiðað nám er eitt sterkasta notkunarsviðið. Mikilvægt: láta einhvern sem kann málið líta yfir þýðingar ef hægt er. (2 mín.)]
]

#slide(title: "3 · Endurgjöf á verkefni")[
  #v(0.4em)
  #grid(columns: (1.15fr, 1fr), column-gutter: 24pt,
    says[„Hér eru ritgerðir bekkjarins og matskvarðinn minn. Skrifaðu drög að endurgjöf fyrir hverja ritgerð og búðu til yfirlit yfir það sem margir þurfa að æfa betur.“],
    does(
      [Les hverja ritgerð],
      [Ber saman við matskvarðann],
      [Skrifar drög að endurgjöf],
      [Tekur saman algeng atriði í töflu],
    ),
  )
  #v(1fr)
  #align(center, block(fill: red-soft, radius: 12pt, inset: 12pt,
    text(size: 19pt, fill: red)[*Þú* ákveður einkunnir og lokaorð. Notaðu aðeins kerfi sem *skólinn hefur samþykkt* fyrir verkefni nemenda.]))
  #notes[Hér þarf að fara varlega: persónuupplýsingar nemenda. Fjarlægja nöfn ef hægt er og fylgja reglum skólans. Yfirlitið yfir algeng atriði nýtist vel til að skipuleggja næstu kennslustund. (2–3 mín.)]
]

#slide(title: "4 · Skipulag og samskipti")[
  #v(0.4em)
  #grid(columns: (1.15fr, 1fr), column-gutter: 24pt,
    says[„Skrifaðu stutt og hlýlegt fréttabréf til foreldra um það sem við gerðum í október, út frá vikuáætlununum mínum. Bættu við dagsetningum á viðburðum í nóvember.“],
    does(
      [Les vikuáætlanirnar],
      [Finnur viðburði í dagatalinu],
      [Skrifar drög og setur upp útlit],
      [Skilar tilbúnu til yfirlestrar],
    ),
  )
  #v(1fr)
  #align(center, text(size: 20pt, fill: gray)[Annað dæmi: „Taktu til í möppunni minni og raðaðu skjölum eftir fögum.“])
  #notes[Stjórnsýslan tekur tíma frá kennslunni. Hér er auðvelt að byrja — lítil áhætta og mikill tímasparnaður. Fréttabréfið fer ekki af stað fyrr en þú hefur lesið það. (2 mín.)]
]

#slide(title: "Mánudagsmorgunn með fulltrúa")[
  #v(0.8em)
  #let moment(time, body, c, soft) = block(fill: soft, radius: 14pt, inset: 14pt, width: 100%, height: 4.6cm)[
    #text(size: 26pt, weight: "bold", fill: c, time)
    #v(-0.2em)
    #text(size: 17pt, body)
  ]
  #grid(columns: (1fr, auto, 1fr, auto, 1fr, auto, 1fr), column-gutter: 6pt, align: horizon,
    moment("8:00", [Þú biður um verkefni vikunnar.], chat-c, chat-soft),
    arrow-r(len: 0.8cm),
    moment("8:10", [Þú ferð og *kennir*.], navy, light),
    arrow-r(len: 0.8cm),
    moment("10:00", [Drögin bíða tilbúin.], agent-c, agent-soft),
    arrow-r(len: 0.8cm),
    moment("10:15", [Þú ferð yfir, lagar og *samþykkir*.], accent, accent-soft),
  )
  #v(1fr)
  #align(center, text(size: 26pt, weight: "bold", fill: navy)[Fulltrúinn vinnur á meðan þú kennir.])
  #notes[Stóri munurinn frá spjalli: þú þarft ekki að sitja við tölvuna á meðan. Þú gefur verkefni og snýrð þér að nemendunum. (2 mín.)]
]

// ============================================================
//  HLUTI 4 — Að vinna vel
// ============================================================
#section(4, "Að vinna vel", "Góð fyrirmæli og skynsamleg notkun")

#slide(title: "Skrifaðu eins og fyrir forfallakennara")[
  #v(0.2em)
  #text(size: 20pt, fill: gray)[Góður forfallakennari þarf að vita fjóra hluti. Fulltrúi líka.]
  #v(0.4em)
  #grid(columns: (1fr, 1fr, 1fr, 1fr), column-gutter: 16pt,
    card(fill: agent-soft, color: agent-c, title: "Markmið")[
      #text(size: 19pt, weight: "bold")[Hvað á að gera?] \ #v(0.2em) „Búðu til krossapróf …“],
    card(fill: chat-soft, color: chat-c, title: "Samhengi")[
      #text(size: 19pt, weight: "bold")[Fyrir hverja?] \ #v(0.2em) „… fyrir 7. bekk sem er nýbúinn með landnámið …“],
    card(fill: accent-soft, color: accent, title: "Útkoma")[
      #text(size: 19pt, weight: "bold")[Hvernig á það að líta út?] \ #v(0.2em) „… 10 spurningar, svör á sér blaði, Word-skjal.“],
    card(fill: red-soft, color: red, title: "Mörk")[
      #text(size: 19pt, weight: "bold")[Hvað má ekki?] \ #v(0.2em) „Ekki senda neitt — leyfðu mér að lesa fyrst.“],
  )
  #notes[Kennarar kunna þetta nú þegar — þeir skrifa leiðbeiningar fyrir forfallakennara reglulega. Sama færni! (3 mín.)]
]

#slide(title: "Óljóst eða skýrt?")[
  #v(0.6em)
  #grid(columns: (1fr, 1fr), column-gutter: 26pt,
    block(fill: red-soft, radius: 16pt, inset: 22pt, width: 100%, height: 8.4cm)[
      #tag(color: red)[ÓLJÓST]
      #v(0.2em)
      #text(size: 26pt, style: "italic")[„Gerðu verkefni um Ísland.“]
      #v(1fr)
      #text(size: 18pt, fill: red)[Fyrir hvaða aldur? Hvaða efni? Hversu langt? Á hvaða formi? Fulltrúinn þarf að giska.]
    ],
    block(fill: agent-soft, radius: 16pt, inset: 22pt, width: 100%, height: 8.4cm)[
      #tag(color: agent-c)[SKÝRT]
      #v(0.2em)
      #text(size: 19pt, style: "italic")[„Búðu til 10 spurninga krossapróf um landnám Íslands fyrir 7. bekk. Notaðu einfalt mál. Settu svörin á sér blaðsíðu og vistaðu sem Word-skjal í möppunni Samfélagsfræði.“]
      #v(1fr)
      #text(size: 18pt, fill: agent-c)[Markmið ✓ Samhengi ✓ Útkoma ✓]
    ],
  )
  #notes[Óljós fyrirmæli gefa óljósa niðurstöðu — alveg eins og með fólk. Ef útkoman er ekki góð: bæta fyrirmælin og reyna aftur. (2 mín.)]
]

#slide(title: "Þú ert enn kennarinn")[
  #v(1em)
  #align(center, flow(
    node(fill: chat-soft, color: chat-c, w: 5cm)[*Þú* \ gefur verkefni],
    arrow-r(),
    node(fill: agent-soft, color: agent-c, w: 5cm)[*Fulltrúinn* \ vinnur drög],
    arrow-r(),
    node(fill: accent-soft, color: accent, w: 5cm)[*Þú* \ ferð yfir og lagar],
    arrow-r(),
    node(fill: navy, color: white, w: 4.4cm)[*Þú* \ ákveður],
  ))
  #v(1fr)
  #big-quote[Fulltrúinn skrifar drög. \ Kennarinn tekur ákvarðanir.]
  #v(1fr)
  #notes[Fullvissa fólk: þetta snýst ekki um að skipta kennurum út. Faglegt mat, þekking á nemendum og tengslin við þá eru alltaf kennarans. Þrjú af fjórum skrefum eru þín. (2 mín.)]
]

#slide(title: "Fjögur atriði til að hafa í huga")[
  #v(0.4em)
  #grid(columns: (1fr, 1fr), column-gutter: 22pt, row-gutter: 18pt,
    card(fill: red-soft, color: red, title: "Persónuvernd")[Settu ekki nöfn eða viðkvæmar upplýsingar um nemendur í kerfi sem skólinn hefur ekki samþykkt.],
    card(fill: accent-soft, color: accent, title: "Villur")[Gervigreind getur haft rangt fyrir sér — og hljómað mjög sannfærandi á meðan. Lestu alltaf yfir.],
    card(fill: chat-soft, color: chat-c, title: "Staðreyndir")[Athugaðu ártöl, nöfn og tölur. Biddu fulltrúann um heimildir.],
    card(fill: light, color: navy, title: "Reglur skólans")[Fylgdu stefnu skólans og sveitarfélagsins um gervigreind.],
  )
  #notes[Ekki hræða fólk — en þetta er mikilvægt. Persónuvernd er stærsta atriðið. Hvetja fólk til að spyrja stjórnendur hvaða lausnir eru samþykktar í þeirra skóla. (3 mín.)]
]

#slide(title: "Byrjaðu smátt")[
  #v(1em)
  #let stepc(n, t, sub, c, soft) = block(fill: soft, radius: 16pt, inset: 18pt, width: 100%, height: 7.6cm)[
    #actor(str(n), fill: c, r: 0.7cm, size: 24pt)
    #v(0.2em)
    #text(size: 23pt, weight: "bold", fill: c, t)
    #v(-0.2em)
    #text(size: 18pt, sub)
  ]
  #grid(columns: (1fr, auto, 1fr, auto, 1fr), column-gutter: 10pt, align: horizon,
    stepc(1, [Veldu eitt verkefni], [Eitthvað endurtekið sem þér leiðist — t.d. vikupóstinn til foreldra.], chat-c, chat-soft),
    arrow-r(len: 1cm),
    stepc(2, [Láttu fulltrúann prófa], [Gefðu skýr fyrirmæli: markmið, samhengi, útkoma, mörk.], agent-c, agent-soft),
    arrow-r(len: 1cm),
    stepc(3, [Farðu yfir og lærðu], [Hvað var gott? Hvað vantaði? Bættu fyrirmælin næst.], accent, accent-soft),
  )
  #v(1fr)
  #align(center, text(size: 24pt, weight: "bold", fill: navy)[Lítil skref — mikill tímasparnaður.])
  #notes[Hvatning: enginn þarf að breyta öllu í einu. Eitt verkefni í næstu viku. (2 mín.)]
]

#slide(title: "Ykkar hugmyndir")[
  #v(1fr)
  #align(center, block(width: 92%, fill: accent-soft, radius: 20pt, inset: 32pt)[
    #text(size: 34pt, weight: "bold", fill: navy)[Hvaða verkefni myndir þú \ vilja afhenda fulltrúa?]
    #v(0.6em)
    #text(size: 22pt)[Ræðið við sessunaut í *3 mínútur*. \ Veljið eitt verkefni og skrifið fyrirmælin saman.]
  ])
  #v(1fr)
  #align(center, text(size: 18pt, fill: gray)[Markmið · Samhengi · Útkoma · Mörk])
  #v(0.6em)
  #notes[Tímataka 3 mínútur. Fá svo 2–3 pör til að deila hugmyndinni sinni og fyrirmælunum. Hrósa og bæta við. (6–8 mín. samtals)]
]

#slide(title: "Það sem við tökum með okkur")[
  #v(0.8em)
  #let point(n, t, c) = block(fill: light, radius: 14pt, inset: 16pt, width: 100%,
    grid(columns: (1.6cm, 1fr), column-gutter: 16pt, align: horizon,
      actor(str(n), fill: c, r: 0.65cm, size: 24pt), text(size: 26pt, t)))
  #stack(spacing: 18pt,
    point(1, [Spjall *svarar* — fulltrúi *framkvæmir*.], agent-c),
    point(2, [Gefðu skýr fyrirmæli — eins og fyrir *forfallakennara*.], chat-c),
    point(3, [*Þú ert alltaf kennarinn*: farðu yfir og ákveddu.], accent),
  )
  #notes[Draga saman þrjá punkta. Ef fólk man bara eitt: spjall svarar, fulltrúi framkvæmir. (2 mín.)]
]

#page(fill: navy, footer: none)[
  #set text(fill: white)
  #align(center + horizon)[
    #text(size: 60pt, weight: "bold")[Takk fyrir!]
    #v(0.2em)
    #text(size: 28pt, fill: accent)[Spurningar?]
  ]
  #notes[Opna fyrir spurningar. Bjóða aðstoð við að prófa fyrsta verkefnið. (Afgangstími)]
]
