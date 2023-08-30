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
		help="datetime in UNIX epoch maximum duplicate data checked for. Default 0 means whatever the maximum data in database")
	parser.add_argument("-rfs", "--drop_first_stop", type = str, default = "false",
		help="whether drop the first stop-trip pair to residuals to be captured by prior data processed later")
	parser.add_argument("-am", "--azimuth_missing", type = str, default = "false",
		help="if azimuth field is not set in parquet or protobuff set to true so yuloserver will not erroneously use the default value 0")
	parser.add_argument("-sm", "--speed_missing", type = str, default = "false",
		help="if speed field is not set in parquet or protobuff set to true so yuloserver will not erroneously use the default value 0")
	parser.add_argument("-raw", "--raw_output", type = str, default = "false",
		help="saves a raw json in working directory ratherthan uploading to database")

	args = parser.parse_args()
	print(args)

	##delete any aborted json files
	loc_files = os.listdir()
	#[os.delete(f) for f in loc_files if f.endswith("json")]

	ul_dir = args.directory
	print(ul_dir)
	files = [f for f in os.listdir(ul_dir) if search(args.filestem, f) and  (f.endswith(".csv") or f.endswith(".gz") or f.endswith(".parquet") or f.endswith(".pbf"))]
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

		if args.azimuth_missing == 'true':
			command = command + "-H 'azimuth_missing: true'"
		if args.speed_missing == 'true':
			command = command + "-H 'speed_missing: true'"
		if args.raw_output == 'true':
			command = command + "-H 'raw_output: true'"
		print(command)
		if args.raw_output == 'true':
			out  = os.popen(command).read()
			fn_json = os.path.basename(fn) + "OUTPUT.json" 
			with open(fn_json, "w") as f:
				f.write(out)
		else:
			os.system(command)

