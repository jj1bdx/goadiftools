#!/bin/sh
# This script outputs a reduced-size ADIF file to stdoud
# from a pre-defined master ADIF log file (as MASTERFILE)
# for the wsjtx_log.adi file of JTDX for dupe checking.
# 
# What this script does is:
# Pick up only FT8/MFSK(including FT4)/JT65 entries,
# then remove unnecessary entries for wsjtx_log.adi,
# then sort in time increase sequence.
MASTERFILE=your_choice_of_log_file.adi
#
goadifgrep -f ${MASTERFILE} mode "(^FT8$|^MFSK$|^JT65$)" | \
	goadifdelf cqz dxcc ituz cont country my_city notes name qth \
	           qslmsg qsl_via iota tx_pwr stx_string srx_string \
		   cnty state | \
	goadiftime
