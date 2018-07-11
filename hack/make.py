#!/usr/bin/env python

import antipackage
from github.appscode.libbuild import libbuild
from subprocess import call
from os.path import expandvars

libbuild.REPO_ROOT = libbuild.GOPATH + '/src/github.com/kube-ci/experiments'

def run_cmd(cmd):
    print(cmd)
    return call([expandvars(cmd)], shell=True, stdin=None, cwd=libbuild.REPO_ROOT)

libbuild.ungroup_go_imports('*.go', 'apis', 'client', 'pkg')
run_cmd('goimports -w *.go apis client pkg')
run_cmd('gofmt -s -w *.go apis client pkg')

run_cmd('go install')
