#!/bin/bash
# Genera msf_modules.json parseando los módulos instalados localmente

MSF_MODULES="/usr/share/metasploit-framework/modules/exploits"
OUTPUT="$HOME/.deyaclaw/msf_modules.json"

echo "Buscando módulos con CVE en $MSF_MODULES..."

python3 << 'PYEOF'
import os, re, json

msf_path = "/usr/share/metasploit-framework/modules/exploits"
output = os.path.expanduser("~/.deyaclaw/msf_modules.json")

cve_re = re.compile(r"'CVE',\s*'(\d{4}-\d+)'")
result = {}

for root, dirs, files in os.walk(msf_path):
    for f in files:
        if not f.endswith(".rb"):
            continue
        full = os.path.join(root, f)
        try:
            content = open(full).read()
        except:
            continue
        cves = cve_re.findall(content)
        if not cves:
            continue
        # ruta relativa del módulo: exploits/linux/samba/...
        rel = os.path.relpath(full, "/usr/share/metasploit-framework/modules")
        module = rel.replace(".rb", "")
        for cve in cves:
            key = f"CVE-{cve}"
            if key not in result:
                result[key] = module

with open(output, "w") as f:
    json.dump(result, f, indent=2)

print(f"✅ {len(result)} CVEs mapeados → {output}")
PYEOF
