graph:
	go mod graph | modgv | sfdp -Tpng -o graph.png
	