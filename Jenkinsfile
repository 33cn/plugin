
pipeline {
    agent any

    environment {
        GOPATH = "${WORKSPACE}"
        PROJ_DIR = "${WORKSPACE}/src/github.com/33cn/plugin"
    }

    options {
        timeout(time: 2,unit: 'HOURS')
        retry(1)
        timestamps()
        gitLabConnection('gitlab33')
        gitlabBuilds(builds: ['check'])
        checkoutToSubdirectory "src/github.com/33cn/plugin"
    }

    stages {
        stage('deploy') {
            steps {
                dir("${PROJ_DIR}"){
                    gitlabCommitStatus(name: 'deploy'){
                        sh 'make build_ci'
                        // sh "cd build && mkdir ${env.BUILD_NUMBER} && cp ci/* ${env.BUILD_NUMBER} -r && cp chain33* Dockerfile* docker* *.sh ${env.BUILD_NUMBER}/ && cd ${env.BUILD_NUMBER}/ && ./docker-compose-pre.sh run ${env.BUILD_NUMBER} all "
                    }
                }
            }


        }
    }

    post {
        always {
            echo 'One way or another, I have finished'
            // clean up our workspace
            deleteDir()
        }

        success {
            echo 'I succeeeded!'
            echo "email user: ${ghprbActualCommitAuthorEmail}"
            script{
                try {
                    mail to: "test",
                         subject: "Successed Pipeline: ${currentBuild.fullDisplayName}",
                         body: "this is success with ${env.BUILD_URL}"
                }
                catch (err){
                    echo err
                }
            }
            echo currentBuild.result

        }

        failure {
            echo 'I failed '
            echo "email user: ${ghprbActualCommitAuthorEmail}"
            script{
                try {
                    mail to: "test",
                         subject: "Failed Pipeline: ${currentBuild.fullDisplayName}",
                         body: "Something is wrong with ${env.BUILD_URL}"
                }catch (err){
                    echo err
                }
            }

            echo currentBuild.result
        }
    }
}
