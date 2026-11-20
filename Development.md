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



## Setup entwicklung:

ich habe jetzt mal das setup fertig gemacht.
hier der flow.

als erstes habe ich alle ui logiken und global assets in ein package mit dem namen ui gegeben.
hier habe ich jetzt 2 emebddings.

- static:  
einmal den static folder. der static folder hält globale css und js files und dinge wie logos usw.

- templates:
templates hält .html files, die global verwendet werden. ein beispiel ist layout.html

beide werden dann via embedding eingebunden.


dann haben ich renderr.go
der renderer.go wir von jedem feature instanziert und packaged die globalen templates gemeinsam mit den feature spezifischen html templates.
damit habe ich hier keine coalision gefahr mehr.

dann habe ich das registry.go  
registry go erlaubt es mir im feature auch eigene js und css datein zu haben, die dann registriert werden.
registry.Handle() ist ausserdem eine factory mir eine funktionierende handlefunc packaged die ich dann später einbinden kann.

Entscheidungen:

 - Renderer. jedes feature instanziert eigenen render.  
 gibt mir mir flexibilität













