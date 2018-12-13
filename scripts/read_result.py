#!/usr/bin/env python
# -*- coding: utf-8 -*-

'''
Too many comments ### in the result
omit them when read
'''

import sys

def rm_comment(file_path):
	with open(file_path) as f:
		lines = f.readlines()
		for line in lines:
			if not line.startswith("###"):
				print(line, end="")


if __name__ == "__main__":
	rm_comment(sys.argv[1])
