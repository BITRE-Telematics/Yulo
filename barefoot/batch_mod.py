#!/usr/bin/env python

#
# Copyright (C) 2016, BMW Car IT GmbH
#
# Author: Sebastian Mattheis <sebastian.mattheis@bmw-carit.de>
#
# Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
# in compliance with the License. You may obtain a copy of the License at
# http://www.apache.org/licenses/LICENSE-2.0 Unless required by applicable law or agreed to in
# writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
# WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific
# language governing permissions and limitations under the License.
#

##modified version of the original batch, changes by Richard Green 2017

__author__ = "sebastian.mattheis@bmw-carit.de"
__copyright__ = "Copyright 2016 BMW Car IT GmbH"
__license__ = "Apache-2.0"

import optparse
import json
import subprocess
import time
import datetime
import random
import os,sys
import yaml
from multiprocessing import Pool
from functools import partial
import argparse


##It would be possible to concatenate the output from each vehicle here to save read write time in postBarefootmerge.r, but not much

##I might save itermediate storage by deleting the input files as I write them, when I become more confident in the process

def bf(f, cfg, done):
    if f not in done: ##It is much faster doing this in function
        print("Processing %s" % f)
    else:
        print("%s is already done" % f)
        return(f)

    outfile = cfg['output'] + "/" + f

    f_path = cfg['input'] + '/' + f

    
    with open(outfile, 'w') as out:
        subprocess.call("cat %s | netcat %s %s" % (f_path, cfg['host'], cfg['port']), shell=True, stdout = out)
    return(f)

if __name__=='__main__':

    parser = argparse.ArgumentParser()
    parser.add_argument("-p", "--partial", type = str, default = 'False',
                        help="check for already done files, for interruptions")
    args = parser.parse_args()

    with open("barefoot.yaml", 'r') as ymlfile:
        cfg = yaml.safe_load(ymlfile)

    if not os.path.exists(cfg['output']):
        os.makedirs(cfg['output'])

    

    files = [f for f in os.listdir(cfg['input']) if f.endswith(".json")]

    

    if eval(args.partial):
        done = os.listdir(cfg['output'])
    else:
        done = []

    bf1 = partial(bf, cfg = cfg, done = done)

    with Pool() as p:
        p.map(bf1, files)

