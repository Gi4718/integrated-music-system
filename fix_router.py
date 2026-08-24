import sys
f = open(sys.argv[1], 'r')
c = f.read()
f.close()
old = 'download.POST(" /verify-metadata-all, verifyMetadataAll)'
new = 'download.POST("/verify-metadata-all", verifyMetadataAll)'
c = c.replace(old, new)
f = open(sys.argv[1], 'w')
f.write(c)
f.close()
print('fixed')
