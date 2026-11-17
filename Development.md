# Entwickung

Im moment arbeite ich an dem go wiring.
ich versuche fragen zu klären wie projekt struktur und split zwischen view und controller.  
also was ein guter architektur split ist.

ki hat mir hier vorgeschlagen nach features zu splitten und einen generellen render zu haben.
der vorteil ist das jedes feature self containing view und logik hat.  
ich muss nur dem renderer die notwendigen bauteile geben, dass er mit den ganzen partials und layout usw umgehen kann.
da bin ich mir aktuell noch nicht ganz sicher wie ich das am besten mache.

im moment schaffe ich es eine homepage zu rendern. 

der flow ist aktuell wie folgt:


| main.go |  ->  | routers.go | internal/feature/home | -> | handler.go |


im file routers habe ich handlers struct:
dieses hält jedes feature bereit.
die feature handler haben einen Connect endpunkt.
dort werden die Routen an den jeweiligen handler gebunden.
aktuell verwenden wir unserrn Renderer noch nicht.



