#!/bin/bash

go build -C ./backend -o ../bookOrganizerServer && ./bookOrganizerServer $@