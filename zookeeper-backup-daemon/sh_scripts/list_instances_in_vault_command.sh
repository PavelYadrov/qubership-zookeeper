#!/bin/sh
set -x

vault=$1

if [ ! -f ${vault}/znodes.zip ];
then
    ls ${vault};
else
    unzip -l ${vault}/znodes.zip | awk '{ if($4 ~ /^[^\/]+\/$/) print substr($4, 1, length($4)-1)}';
fi