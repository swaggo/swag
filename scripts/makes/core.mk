
ALL_MK_DIR := $(dir $(lastword $(MAKEFILE_LIST)))

include $(ALL_MK_DIR)install.mk
include $(ALL_MK_DIR)utilities.mk
include $(ALL_MK_DIR)github.mk
include $(ALL_MK_DIR)standard.mk