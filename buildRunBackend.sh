#!/bin/bash

go build -C ./backend -o ../out && ./out $@