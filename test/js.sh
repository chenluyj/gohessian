#rm -f *.class
cp="./:./src/"
for j in $(ls lib) 
do
  cp+=":./lib/$j"
done

JAVA_OPT=" -DDEBUG -Dorg.eclipse.jetty.util.log.class=org.eclipse.jetty.util.log.Slf4Log "
JAVA_OPT=" $JAVA_OPT -Dorg.eclipse.jetty.LEVEL=INFO -DServiceImpl.LEVEL=DEBUG -DService.LEVEL=DEBUG "

javac -classpath $cp ./src/HessianServer.java
java $JAVA_OPT -classpath $cp HessianServer
