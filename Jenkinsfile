@Library('laiye') _

pipeline {
    environment {
        PROJECT = 'api-test'
    }
    agent any
    options {
        buildDiscarder logRotator(artifactDaysToKeepStr: '', artifactNumToKeepStr: '', daysToKeepStr: '30', numToKeepStr: '3')
    }
    stages {
//         stage('Tag Controller') {
//             when { tag "v*" }
//             steps {
//                 script {
//                     backend.tag_controller("${TAG_NAME}")
//                 }
//             }
//         }
        stage('Build Docker Images') {
            failFast true
            parallel {
                

                stage('process: api-test-siber') {
                    when { tag "v*" }
                    agent any
                    stages {
                        stage('Prod') {
                            steps {
                                script {
                                    backend.debug_handler()
                                }
                            }
                        }
                        stage('Build api-test-siber image') {
                            steps {
                                echo "api-test-siber"
                                script {
                                    backend.build_image("${env.PROJECT}", "api-test-siber","${env.BRANCH_NAME}", "./docker/siber.Dockerfile")
                                }
                            }
                        }
                    }
                }

                 
                stage('deploy api-test-siber to saas-test') {
                    agent any
                    when { branch "tes*" }
                    steps {
                        script {
                            backend.saas_test_deploy("${env.PROJECT}", "api-test-siber", "./docker/siber.Dockerfile")
                        }
                    }
                }

                stage('deploy api-test-siber to test env') {
                    agent any
                    when { 
                        not { tag "v*" }
                        // not { branch "tes*" }
                    }
                    steps {
                        script {
                            backend.test_env_deploy_with_kube_img("${env.PROJECT}", "api-test-siber", "./docker/siber.Dockerfile")
                        }
                    }
                }
                
            }
        }
    }
    post {
        success {
            script {
                post.post_cibot()
            }
        }
    }
}