package com.studycollab.file_service.grpc;

import com.studycollab.file_service.service.MinioFileService;
import com.studycollab.file_service.proto.DownloadRequest;
import com.studycollab.file_service.proto.DownloadResponse;
import com.studycollab.file_service.proto.FileServiceGrpc;
import com.studycollab.file_service.proto.UploadRequest;
import com.studycollab.file_service.proto.UploadResponse;
import io.grpc.Status; 
import io.grpc.stub.StreamObserver;
import org.lognet.springboot.grpc.GRpcService; 
import org.springframework.beans.factory.annotation.Autowired;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

@GRpcService
public class FileServiceGrpcImpl extends FileServiceGrpc.FileServiceImplBase {
    private static final Logger log = LoggerFactory.getLogger(FileServiceGrpcImpl.class);

    @Autowired
    private MinioFileService minioService;

    @Override
    public void generateUploadUrl(UploadRequest req, StreamObserver<UploadResponse> resp) {
        try {
            log.info("Received upload request for: {}", req.getFilename());
            
            UploadResponse response = minioService.generateUpload(req);
            
            resp.onNext(response);
            resp.onCompleted();

        } catch (Exception e) {
            log.error("Error generating upload URL", e);
            
            
            resp.onError(Status.INTERNAL
                .withDescription("Server error: " + e.getMessage())
                .withCause(e)
                .asRuntimeException());
        }
    }

    @Override
    public void generateDownloadUrl(DownloadRequest req, StreamObserver<DownloadResponse> resp) {
        try {
            DownloadResponse response = minioService.generateDownload(req);
            resp.onNext(response);
            resp.onCompleted();
        } catch (Exception e) {
            log.error("Error generating download URL", e);
            resp.onError(Status.INTERNAL
                .withDescription("Server error: " + e.getMessage())
                .asRuntimeException());
        }
    }
}
