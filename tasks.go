package main

func StartTask() {
	ConnectDatabase()
	TaskIndividualEmails2Domains()
	TaskCertificate2Domain()
	TaskDiscoursePosts2Domains()
	TaskHttpServices2Discourses()
}
