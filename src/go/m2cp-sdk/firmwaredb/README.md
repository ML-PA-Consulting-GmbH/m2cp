# Firmware DB

A filesystem based db for firmware blobs metadata.

Firmware files are stored at:

    <basepath>/<fwt>/<hwr>/<fwr>/   

Each firmware consists of the following files:

- meta.json 
- manifest
- firmware0
- firmware1

The file meta.json contains meta-data describing a firmware, which 
consists of a manifest and two binary firmware files.

The meta.json contains hashes for the other files and policies, for 
which sensors the firmware is allowed to be deployed.

