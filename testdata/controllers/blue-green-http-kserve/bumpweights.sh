echo "kubectl annotate --overwrite isvc wisdom-primary iter8.tools/weight='20'"
echo "kubectl annotate --overwrite isvc wisdom-candidate iter8.tools/weight='80'"
kubectl annotate --overwrite isvc wisdom-primary iter8.tools/weight='20'
kubectl annotate --overwrite isvc wisdom-candidate iter8.tools/weight='80'
