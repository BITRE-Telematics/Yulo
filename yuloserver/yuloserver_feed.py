import os
import argparse
from re import search

if __name__ == "__main__":

	parser = argparse.ArgumentParser()
	parser.add_argument("-s", "--server", type = str, default = "http://0.0.0.0:6969/process",
		help="Yuloserver url and port")
	parser.add_argument("-f", "--filestem", type = str, default = "DUCKSOUPRADIO",
		help="file stem for files to be processed, accepts regex")
	parser.add_argument("-d", "--directory", type = str, default = "../data/",
		help="file directory")
	parser.add_argument("-r", "--gen_resids_only", type = str, default = "false",
		help="whether to only generate and save residuals - for fixing residuals")
	parser.add_argument("-pd", "--prune_dupes", type = str, default = "true",
		help="whether to exclude observations after data already in database")
	parser.add_argument("-mp", "--max_prune", type = int, default = 0,
		help="datetime in UNIX epoch maximum duplicate data allowed. Default 0 means whatever the maximum data in database")
	parser.add_argument("-rfs", "--drop_first_stop", type = str, default = "false",
		help="whether drop the first stop-trip pair to residuals to be captured by prior data processed later")

	args = parser.parse_args()
	print(args)

	##delete any aborted json files
	loc_files = os.listdir()
	#[os.delete(f) for f in loc_files if f.endswith("json")]

	ul_dir = args.directory
	print(ul_dir)
	files = [f for f in os.listdir(ul_dir) if search(args.filestem, f) and  (f.endswith(".csv") or (f.endswith(".gz")))]
	files.sort()


	for fn in files:
		print("File %s" % fn)
		
		command = "curl -X POST -F 'myFile=@%s' %s " % (ul_dir + fn, args.server)
		if args.gen_resids_only == 'true':
			command = command + "-H 'gen_resids_only: true'"
		if args.prune_dupes == 'true':
			command = command + " -H 'prune_dupes: true' -H 'max_prune: %s'" % args.max_prune
		else:
			command = command + " -H 'prune_dupes: false' "

		if args.drop_first_stop == 'true':
			command = command + " -H 'drop_first_stop: true' "
		else:
			command = command + " -H 'drop_first_stop: false' "
		print(command)
		os.system(command)

