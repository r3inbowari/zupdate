taskkill /f /pid 35352
del hola.exe
rename hola.exe_tmp hola.exe
start "hola" hola.exe -a
exit
