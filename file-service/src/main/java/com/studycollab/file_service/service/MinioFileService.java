package com.studycollab.file_service.service;

import com.studycollab.file_service.proto.DownloadRequest;
import com.studycollab.file_service.proto.DownloadResponse;
import com.studycollab.file_service.proto.UploadRequest;
import com.studycollab.file_service.proto.UploadResponse;
import io.minio.GetPresignedObjectUrlArgs;
import io.minio.MinioClient;
import io.minio.http.Method;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.util.UUID;

@Service
public class MinioFileService {

    @Autowired
    private MinioClient client;

    @Value("${minio.bucket}")
    private String bucket;

    public UploadResponse generateUpload(UploadRequest req) {
        String fileId = UUID.randomUUID().toString();

        String presigned;
        try {
            presigned = client.getPresignedObjectUrl(
                GetPresignedObjectUrlArgs.builder()
                    .method(Method.PUT)
                    .bucket(bucket)
                    .object(fileId)
                    .build()
            );
        } catch (Exception e) {
            throw new RuntimeException(e);
        }

        return UploadResponse.newBuilder()
                .setPresignedUrl(presigned)
                .setFileId(fileId)
                .build();
    }

    public DownloadResponse generateDownload(DownloadRequest req) {
        String presigned;
        try {
            presigned = client.getPresignedObjectUrl(
                GetPresignedObjectUrlArgs.builder()
                    .method(Method.GET)
                    .bucket(bucket)
                    .object(req.getFileId())
                    .build()
            );
        } catch (Exception e) {
            throw new RuntimeException(e);
        }

        return DownloadResponse.newBuilder()
                .setPresignedUrl(presigned)
                .build();
    }
}
