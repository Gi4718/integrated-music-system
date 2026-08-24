import sys
f = open(sys.argv[1])
lines = f.readlines()
f.close()
for i, line in enumerate(lines):
    if 'verify-metadata' in line and 'verifyMetadata' in line and 'verify-metadata-all' not in line:
        indent = line[:len(line) - len(line.lstrip())]
        lines.insert(i+1, indent + 'download.POST("/verify-metadata-all", verifyMetadataAll)\n')
        break
f = open(sys.argv[1], 'w')
f.writelines(lines)
f.close()
print('route added')
