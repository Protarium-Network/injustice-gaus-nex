import sys; sys.argv=['x']
exec(open('refs.py').read().split('for s in sys.argv[1:]')[0])
k='65e9f4d6'.encode('utf-16-be')
for a in find_str(k):
    print('key at',hex(a),[hex(x) for x in xrefs(a)])
