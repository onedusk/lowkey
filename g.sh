#!/bin/bash

# echo "# lowkey" >> README.md
git init;
git add -A;
git commit -m "first commit";
git branch -M main;
git remote add origin git@github.com:onedusk/lowkey.git;
git push -u origin main;
