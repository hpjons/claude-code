# Skrifað 13.5.2022
# Helgi Páll Jónsson
# Forritunarmál: Julia


# Forrit sem prentar út vegalengd með tíma í
# frjálsu falli. Ef vegalengd fer yfir 1000
# skrifar forritið út km í stað m.

println("Gefðu upp þyngdarhröðun (m/s^2) : ")
g = readline()
g = parse(Float64, g)

println("Tími (s)\tVegalengd")

for t in 0:100
   s = 0.5 * g * t^2                            # Jafnan s = 0.5 * g * t^2 reiknar vegalengd með tíma í frjálsu falli
      if s > 1000                               # Úttak á skjáinn breytist í km ef farið er yfir 1000 m
         s = (0.5 * g * t^2)/1000.0
         println(t,"\t", round(s; digits=2), " km")
      else
         println(t,"\t", round(s; digits=2), " m")
   end
end
