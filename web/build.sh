cd src
rm -f src.js
rm -f src.js.map
gopherjs build -m
mv src.js ../assets/src.js
mv src.js.map ../assets/src.js.map
