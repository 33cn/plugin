#!/bin/bash



# 在 plugin/plugin_type/plugin_name 找出fork
function subdir_forks() {
	plugin_dir=$1
	plugin_name=$2

	full_dir=$1

	forks=$(grep types.RegisterDappFork "${full_dir}" -R | cut -d '(' -f 2 | cut -d ')' -f 1 | sed 's/ //g')

        if [ -z "${forks}" ]; then
		return
	fi
	
	cnt=$(echo "${forks}" | grep "^\"" | wc -l)
	if [ $cnt -gt 0 ]; then
		name=$(echo $forks | head -n1 | cut -d ',' -f 1 | sed 's/"//g')
		echo "[fork.sub.${name}]"
	else
		echo "[fork.sub.${plugin_name}]";
	fi

	for fork in "${forks}"
	do
		echo "${fork}" | awk -F ',' '{ \
					if(match($2,"\"")) gsub("\"","",$2); else gsub("X$","",$2); \
				    	print $2 "=" $3}'
					#/*print "debug" $1 $2 $3;*/ \

	done
	echo 
}


dir=$(go list -f '{{.Dir}}' github.com/33cn/plugin)/plugin/
plugins=$(find $dir -maxdepth 2 -mindepth 2 -type d | sort)
for plugin in ${plugins}
do 
	name=$(echo $plugin | sed 's/.*\///g')
	subdir_forks $plugin $name
done
