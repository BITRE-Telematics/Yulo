'''
This is a script made necessary by the contraints of departmental IT and neo4j. It merely transfers the contents of UploadData to the database server via curl. I could zip it in the future
'''
import os
import argparse
from re import search

if __name__ == "__main__":



	parser = argparse.ArgumentParser()
	parser.add_argument("-k", "--key", type = str, default = "file_transfer_key.txt",
		help="key for upload server")
	parser.add_argument("-s", "--server", type = str, default = "http://141.120.98.192:1337/upload",
		help="upload server URL")
	parser.add_argument("-f", "--filestem", type = str, default = "",
		help="file stem for files to be transfered, accepts regex")
	parser.add_argument("-d", "--directory", type = str, default = "../UploadData/",
		help="file directory. Also changes upload location on server side")
	parser.add_argument("-rm", "--remove", type = str, default = "false",
		help="whether to delete files")

	args = parser.parse_args()


	with open(args.key, 'r') as f:
		key = f.read().splitlines()[0] 
	
	ul_dir = args.directory

	if ul_dir[-1] != '/': ul_dir = ul_dir + '/'
	print(ul_dir)
	files = [f for f in os.listdir(ul_dir) if search(args.filestem, f)]

	header_other = "-H 'other: false'" if ul_dir == "../UploadData/" else "-H 'other: true'"
	
	for fn in files:
		print("File %s" % fn)
		header_file = "-H 'filename: %s'" % fn
		if args.remove == "true":
			command = "curl -X POST %s -H 'key: %s' %s %s  -H 'delete: true'" % (args.server, key, header_other, header_file) 
		else:
			command = "curl -X POST -F 'myFile=@%s' %s -H 'key: %s' %s %s" % (ul_dir + fn, args.server, key, header_other, header_file)
		print(command)
		os.system(command)
