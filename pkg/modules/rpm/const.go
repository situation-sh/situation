//go:build linux

package rpm

const ( //                          Size 						Alignment
	RPM_NULL_TYPE         = 0 //	Not Implemented.
	RPM_CHAR_TYPE         = 1 //	1							1
	RPM_INT8_TYPE         = 2 //	1							1
	RPM_INT16_TYPE        = 3 //	2							2
	RPM_INT32_TYPE        = 4 //	4							4
	RPM_INT64_TYPE        = 5 // 	Reserved.
	RPM_STRING_TYPE       = 6 //	variable, NUL terminated	1
	RPM_BIN_TYPE          = 7 //	1							1
	RPM_STRING_ARRAY_TYPE = 8 //	Variable, sequence of NUL terminated strings	1
	RPM_I18NSTRING_TYPE   = 9 //	variable, sequence of NUL terminated strings	1
)

const ( //                 Tag Value   Type		Count	Status
	RPMTAG_NAME              = 1000 // STRING		1	Required
	RPMTAG_VERSION           = 1001 // STRING		1	Required
	RPMTAG_RELEASE           = 1002 // STRING		1	Required
	RPMTAG_SUMMARY           = 1004 // I18NSTRING	1	Required
	RPMTAG_DESCRIPTION       = 1005 // I18NSTRING	1	Required
	RPMTAG_SIZE              = 1009 // INT32		1	Required
	RPMTAG_DISTRIBUTION      = 1010 // STRING		1	Informational
	RPMTAG_VENDOR            = 1011 // STRING		1	Informational
	RPMTAG_LICENSE           = 1014 // STRING		1	Required
	RPMTAG_PACKAGER          = 1015 // STRING		1	Informational
	RPMTAG_GROUP             = 1016 // I18NSTRING	1	Required
	RPMTAG_URL               = 1020 // STRING		1	Informational
	RPMTAG_OS                = 1021 // STRING		1	Required
	RPMTAG_ARCH              = 1022 // STRING		1	Required
	RPMTAG_SOURCERPM         = 1044 // STRING		1	Informational
	RPMTAG_ARCHIVESIZE       = 1046 // INT32		1	Optional
	RPMTAG_RPMVERSION        = 1064 // STRING		1	Informational
	RPMTAG_COOKIE            = 1094 // STRING		1	Optional
	RPMTAG_DISTURL           = 1123 // STRING		1	Informational
	RPMTAG_PAYLOADFORMAT     = 1124 // STRING		1	Required
	RPMTAG_PAYLOADCOMPRESSOR = 1125 // STRING		1	Required
	RPMTAG_PAYLOADFLAGS      = 1126 // STRING		1	Required
	RPMTAG_OLDFILENAMES      = 1027 // STRING_ARRAY	�	Optional
	RPMTAG_FILESIZES         = 1028 // INT32		�	Required
	RPMTAG_FILEMODES         = 1030 // INT16		�	Required
	RPMTAG_FILERDEVS         = 1033 // INT16		�	Required
	RPMTAG_FILEMTIMES        = 1034 // INT32		�	Required
	RPMTAG_FILEMD5S          = 1035 // STRING_ARRAY	�	Required
	RPMTAG_FILELINKTOS       = 1036 // STRING_ARRAY	�	Required
	RPMTAG_FILEFLAGS         = 1037 // INT32		�	Required
	RPMTAG_FILEUSERNAME      = 1039 // STRING_ARRAY	�	Required
	RPMTAG_FILEGROUPNAME     = 1040 // STRING_ARRAY	�	Required
	RPMTAG_FILEDEVICES       = 1095 // INT32		�	Required
	RPMTAG_FILEINODES        = 1096 // INT32		�	Required
	RPMTAG_FILELANGS         = 1097 // STRING_ARRAY	�	Required
	RPMTAG_DIRINDEXES        = 1116 // INT32		�	Optional
	RPMTAG_BASENAMES         = 1117 // STRING_ARRAY	�	Optional
	RPMTAG_DIRNAMES          = 1118 // STRING_ARRAY	�	Optional
)
