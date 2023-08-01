ARG base_image
FROM ${base_image}

# Set required CNB information
ARG stack_id
LABEL io.buildpacks.stack.id="${stack_id}"

# Set user and group (as declared in base image)
USER ${CNB_USER_ID}:${CNB_GROUP_ID}